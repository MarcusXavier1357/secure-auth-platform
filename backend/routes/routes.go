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
	requireAnyPerm := func(codes ...string) fiber.Handler {
		return middleware.RequireAnyPermission(d.PermService, codes...)
	}

	// Públicas
	app.Post("/auth/login", d.AuthHandler.Login)
	app.Post("/auth/refresh", d.AuthHandler.Refresh)
	app.Post("/auth/logout", d.AuthHandler.Logout)

	// Autenticadas
	app.Get("/me", authMW, d.UserHandler.Me)
	app.Get("/auth/sessions", authMW, d.AuthHandler.ListSessions)
	app.Delete("/auth/sessions", authMW, d.AuthHandler.RevokeAllSessionsExceptCurrent)
	app.Delete("/auth/sessions/:id", authMW, d.AuthHandler.RevokeSession)

	// Gestão de usuários
	app.Get("/users", authMW, requirePerm("users.read"), d.UserHandler.List)
	app.Post("/users", authMW, requirePerm("users.create"), d.UserHandler.Create)
	app.Get("/users/:id", authMW, requirePerm("users.read"), d.UserHandler.Get)
	app.Patch("/users/:id", authMW, requireAnyPerm(
		"users.update", "users.password.reset", "users.deactivate",
	), d.UserHandler.Update)

	app.Get("/roles", authMW, requirePerm("users.read"), d.PermHandler.ListRoles)

	// Gestão de permissões
	app.Get("/permissions", authMW, requirePerm("permissions.read"), d.PermHandler.List)
	app.Post("/permissions", authMW, requirePerm("permissions.create"), d.PermHandler.Create)
	app.Patch("/permissions/:id", authMW, requirePerm("permissions.update"), d.PermHandler.Update)
	app.Delete("/permissions/:id", authMW, requirePerm("permissions.delete"), d.PermHandler.Delete)
	app.Post("/users/:id/permissions", authMW, requirePerm("permissions.grant"), d.PermHandler.Grant)
	app.Delete("/users/:id/permissions/:permissionId", authMW, requirePerm("permissions.revoke"), d.PermHandler.Revoke)

	// Auditoria
	app.Get("/audit-logs", authMW, requirePerm("audit_logs.read"), d.AuditHandler.List)
}
