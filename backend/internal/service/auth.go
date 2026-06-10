package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"auth-backend/cache"
	"auth-backend/internal/models"
	"auth-backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRateLimited        = errors.New("rate limited")
	ErrRateLimitDown      = errors.New("rate limiter unavailable")
	ErrInvalidSession     = errors.New("invalid session")
)

// dummyPasswordHash é comparado nos caminhos de falha sem usuário (inexistente
// ou inativo) para igualar o tempo de resposta ao caminho com bcrypt real —
// sem isso, a diferença de timing permite enumerar emails cadastrados.
var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("timing-equalizer-dummy"), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("generating dummy bcrypt hash: %v", err))
	}
	return hash
}()

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type AuthService struct {
	users       *repository.UserRepository
	sessions    *repository.SessionRepository
	rateLimiter *cache.RateLimiter
	audit       *AuditService

	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(
	users *repository.UserRepository,
	sessions *repository.SessionRepository,
	rateLimiter *cache.RateLimiter,
	audit *AuditService,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:       users,
		sessions:    sessions,
		rateLimiter: rateLimiter,
		audit:       audit,
		jwtSecret:   []byte(jwtSecret),
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}
}

// Login aplica rate limit (antes do bcrypt), valida credenciais, cria a
// sessão no PostgreSQL e retorna o par de tokens.
func (s *AuthService) Login(ctx context.Context, email, password, ip string) (*TokenPair, error) {
	ipKey := "ratelimit:login:ip:" + ip
	emailKey := "ratelimit:login:email:" + email

	ipCount, ipErr := s.rateLimiter.Increment(ctx, ipKey)
	emailCount, emailErr := s.rateLimiter.Increment(ctx, emailKey)
	if ipErr != nil || emailErr != nil {
		// Redis indisponível: rejeitar com 503 para não abrir brecha de brute force.
		slog.Error("rate limiter unavailable during login",
			"ipError", ipErr, "emailError", emailErr)
		return nil, ErrRateLimitDown
	}
	if s.rateLimiter.Exceeded(ipCount) || s.rateLimiter.Exceeded(emailCount) {
		s.audit.Log(ctx, nil, "login.failed", "auth", nil, nil,
			map[string]any{"email": email, "ip": ip, "reason": "rate_limited"})
		return nil, ErrRateLimited
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
			s.auditLoginFailed(ctx, email, ip, "user_not_found")
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.Active {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
		s.auditLoginFailed(ctx, email, ip, "user_inactive")
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.auditLoginFailed(ctx, email, ip, "wrong_password")
		return nil, ErrInvalidCredentials
	}

	pair, err := s.createSession(ctx, user)
	if err != nil {
		return nil, err
	}

	// Login OK limpa o contador por email; o de IP permanece (janela expira sozinha).
	if err := s.rateLimiter.Reset(ctx, emailKey); err != nil {
		slog.Warn("failed to reset login rate limit", "email", email, "error", err)
	}

	s.audit.Log(ctx, &user.ID, "login.success", "auth", &user.ID, nil,
		map[string]any{"email": email, "ip": ip})

	return pair, nil
}

// Refresh valida a sessão pelo hash do refresh token, rotaciona o token
// (invalida o anterior na mesma sessão) e emite novo access token.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	session, err := s.sessions.FindActiveByTokenHash(ctx, hashToken(refreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil || !user.Active {
		return nil, ErrInvalidSession
	}

	newRefresh, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	newExpiry := time.Now().Add(s.refreshTTL)

	if err := s.sessions.RotateToken(ctx, session.ID, hashToken(newRefresh), newExpiry); err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     newRefresh,
		RefreshExpiresAt: newExpiry,
	}, nil
}

// Logout revoga a sessão associada ao refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	session, err := s.sessions.FindActiveByTokenHash(ctx, hashToken(refreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // sessão já inválida; logout é idempotente
		}
		return err
	}

	if err := s.sessions.Revoke(ctx, session.ID); err != nil {
		return err
	}

	s.audit.Log(ctx, &session.UserID, "logout", "auth", &session.UserID, nil, nil)
	return nil
}

// ValidateAccessToken valida assinatura e expiração do JWT e retorna o userId.
func (s *AuthService) ValidateAccessToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("invalid subject")
	}
	return int64(sub), nil
}

func (s *AuthService) createSession(ctx context.Context, user *models.User) (*TokenPair, error) {
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(s.refreshTTL)

	session := &models.Session{
		UserID:           user.ID,
		RefreshTokenHash: hashToken(refreshToken),
		ExpiresAt:        expiresAt,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: expiresAt,
	}, nil
}

// generateAccessToken emite o JWT. Permissões não entram no payload —
// são sempre consultadas via Redis/PostgreSQL.
func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(s.accessTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func (s *AuthService) auditLoginFailed(ctx context.Context, email, ip, reason string) {
	s.audit.Log(ctx, nil, "login.failed", "auth", nil, nil,
		map[string]any{"email": email, "ip": ip, "reason": reason})
}

func generateRefreshToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// hashToken usa SHA-256 — permite lookup direto pelo hash, ao contrário do
// bcrypt. O token tem 256 bits de entropia, então rainbow tables não se aplicam.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
