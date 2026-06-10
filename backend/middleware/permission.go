package middleware

import (
	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/service"
)

// RequirePermission garante que o usuário autenticado possui a permissão.
// Fluxo: Redis (cache-aside) → PostgreSQL no miss → 403 se não possuir.
func RequirePermission(perms *service.PermissionService, code string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := UserID(c)
		if userID == 0 {
			return fiber.NewError(fiber.StatusUnauthorized, "not authenticated")
		}

		allowed, err := perms.HasPermission(c.Context(), userID, code)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "permission check failed")
		}
		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, "missing permission: "+code)
		}
		return c.Next()
	}
}
