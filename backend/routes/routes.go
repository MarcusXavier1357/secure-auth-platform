package routes

import (
	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/audit"
	"auth-backend/internal/auth"
	"auth-backend/internal/permissions"
	"auth-backend/internal/repository"
	"auth-backend/internal/service"
	"auth-backend/internal/users"
	"auth-backend/middleware"
)

type Deps struct {
	AuthHandler  *auth.Handler
	UserHandler  *users.Handler
	PermHandler  *permissions.Handler
	AuditHandler *audit.Handler

	AuthService *service.AuthService
	PermService *service.PermissionService
	UserRepo    *repository.UserRepository
}

func Setup(app *fiber.App, d Deps) {
	authMW := middleware.Auth(d.AuthService, d.UserRepo)
	requirePerm := func(code string) fiber.Handler {
		return middleware.RequirePermission(d.PermService, code)
	}

	// Públicas
	app.Post("/auth/login", d.AuthHandler.Login)
	app.Post("/auth/refresh", d.AuthHandler.Refresh)
	app.Post("/auth/logout", d.AuthHandler.Logout)

	// Autenticadas
	app.Get("/me", authMW, d.UserHandler.Me)

	// Gestão de usuários
	app.Get("/users", authMW, requirePerm("users.manage"), d.UserHandler.List)
	app.Post("/users", authMW, requirePerm("users.manage"), d.UserHandler.Create)
	app.Get("/users/:id", authMW, requirePerm("users.manage"), d.UserHandler.Get)
	app.Patch("/users/:id", authMW, requirePerm("users.manage"), d.UserHandler.Update)

	// Gestão de permissões
	app.Get("/permissions", authMW, requirePerm("permissions.manage"), d.PermHandler.List)
	app.Post("/users/:id/permissions", authMW, requirePerm("permissions.manage"), d.PermHandler.Grant)
	app.Delete("/users/:id/permissions/:permissionId", authMW, requirePerm("permissions.manage"), d.PermHandler.Revoke)

	// Auditoria
	app.Get("/audit-logs", authMW, requirePerm("audit_logs.read"), d.AuditHandler.List)
}
