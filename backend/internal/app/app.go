// Package app monta a aplicação completa (migrations, conexões, wiring de
// services e rotas). É usado pelo main e pelos testes ponta a ponta.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"

	"auth-backend/cache"
	"auth-backend/database"
	"auth-backend/internal/audit"
	"auth-backend/internal/auth"
	"auth-backend/internal/models"
	"auth-backend/internal/permissions"
	"auth-backend/internal/repository"
	"auth-backend/internal/service"
	"auth-backend/internal/users"
	"auth-backend/routes"
)

type Config struct {
	DatabaseURL    string
	MigrationsPath string

	RedisURL       string
	RedisKeyPrefix string

	JWTPrivateKeyPath  string
	JWTPublicKeyPath   string
	JWTKeyPair         *service.JWTKeyPair // testes: par em memória
	AccessTTL          time.Duration
	RefreshTTL         time.Duration

	PermCacheTTL time.Duration

	// Tiers de rate limit de login (plano2 fase 4).
	LoginRateTiers  []cache.RateTier
	LoginCounterTTL time.Duration

	AdminEmail    string
	AdminPassword string

	CookieSecure bool
}

// Retenção e frequência da limpeza de sessões expiradas/revogadas.
const (
	sessionCleanupInterval = time.Hour
	sessionRetention       = 7 * 24 * time.Hour
)

// DefaultLoginRateTiers conforme plano2.md.
func DefaultLoginRateTiers() []cache.RateTier {
	return []cache.RateTier{
		{Threshold: 5, Block: time.Minute},
		{Threshold: 10, Block: 15 * time.Minute},
		{Threshold: 20, Block: 24 * time.Hour},
	}
}

type App struct {
	Fiber *fiber.App
	DB    *bun.DB
	Redis *cache.Client

	cfg         Config
	userRepo    *repository.UserRepository
	permRepo    *repository.PermissionRepository
	sessionRepo *repository.SessionRepository
	auditSvc    *service.AuditService
	stopCleanup chan struct{}
}

func New(cfg Config) (*App, error) {
	if err := database.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	redisClient, err := cache.NewClient(cfg.RedisURL, cfg.RedisKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}
	if err := redisClient.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	var jwtKeys *service.JWTKeyPair
	switch {
	case cfg.JWTKeyPair != nil:
		jwtKeys = cfg.JWTKeyPair
	default:
		var err error
		jwtKeys, err = service.LoadJWTKeys(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("jwt keys: %w", err)
		}
	}

	tiers := cfg.LoginRateTiers
	if len(tiers) == 0 {
		tiers = DefaultLoginRateTiers()
	}
	counterTTL := cfg.LoginCounterTTL
	if counterTTL == 0 {
		counterTTL = 24 * time.Hour
	}

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	permCache := cache.NewPermissionCache(redisClient, cfg.PermCacheTTL)
	tieredLimiter := cache.NewTieredRateLimiter(redisClient, tiers, counterTTL)

	auditSvc := service.NewAuditService(auditRepo)
	permSvc := service.NewPermissionService(permRepo, permCache, auditSvc)
	authSvc := service.NewAuthService(userRepo, sessionRepo, tieredLimiter, auditSvc, jwtKeys, cfg.AccessTTL, cfg.RefreshTTL)
	userSvc := service.NewUserService(userRepo, sessionRepo, permSvc, auditSvc)

	fiberApp := fiber.New(fiber.Config{
		AppName:      "auth-api",
		ErrorHandler: errorHandler,
	})

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Ping(); err != nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "postgres unavailable")
		}
		if err := redisClient.Ping(c.Context()); err != nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "redis unavailable")
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	routes.Setup(fiberApp, routes.Deps{
		AuthHandler:  auth.NewHandler(authSvc, cfg.CookieSecure, 15*time.Minute),
		UserHandler:  users.NewHandler(userSvc, permSvc),
		PermHandler:  permissions.NewHandler(permSvc),
		AuditHandler: audit.NewHandler(auditSvc),
		AuthService:  authSvc,
		PermService:  permSvc,
		UserRepo:     userRepo,
	})

	application := &App{
		Fiber:       fiberApp,
		DB:          db,
		Redis:       redisClient,
		cfg:         cfg,
		userRepo:    userRepo,
		permRepo:    permRepo,
		sessionRepo: sessionRepo,
		auditSvc:    auditSvc,
		stopCleanup: make(chan struct{}),
	}
	go application.runSessionCleanup()

	return application, nil
}

func errorHandler(c *fiber.Ctx, err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
	}

	slog.Error("unhandled request error",
		"method", c.Method(), "path", c.Path(), "error", err)
	return c.Status(fiber.StatusInternalServerError).
		JSON(fiber.Map{"error": "internal server error"})
}

func (a *App) runSessionCleanup() {
	ticker := time.NewTicker(sessionCleanupInterval)
	defer ticker.Stop()

	for {
		a.cleanupSessions()
		select {
		case <-a.stopCleanup:
			return
		case <-ticker.C:
		}
	}
}

func (a *App) cleanupSessions() {
	deleted, expired, err := a.sessionRepo.DeleteExpired(context.Background(), sessionRetention)
	if err != nil {
		slog.Error("session cleanup failed", "error", err)
		return
	}
	for _, sess := range expired {
		sid := sess.ID
		a.auditSvc.Log(context.Background(), &sess.UserID, "session.expired", "session", &sid, nil,
			map[string]any{"sessionId": sess.ID, "reason": "cleanup"})
	}
	if deleted > 0 {
		slog.Info("expired sessions deleted", "count", deleted)
	}
}

func (a *App) SeedAdmin(ctx context.Context) error {
	count, err := a.userRepo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	if a.cfg.AdminPassword == "" {
		return errors.New("ADMIN_PASSWORD is required to seed the first admin user")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(a.cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var adminRoleID *int64
	role := new(models.Role)
	if err := a.DB.NewSelect().Model(role).Where("name = ?", "Admin").Scan(ctx); err == nil {
		adminRoleID = &role.ID
	}

	admin := &models.User{
		Name:         "Administrador",
		Email:        a.cfg.AdminEmail,
		PasswordHash: string(hash),
		RoleID:       adminRoleID,
		Active:       true,
	}
	if err := a.userRepo.Create(ctx, admin); err != nil {
		return err
	}
	if err := a.permRepo.GrantAll(ctx, admin.ID); err != nil {
		return err
	}

	slog.Info("admin user seeded", "email", a.cfg.AdminEmail)
	return nil
}

func (a *App) Close() {
	close(a.stopCleanup)
	if err := a.DB.Close(); err != nil {
		slog.Warn("closing db", "error", err)
	}
}
