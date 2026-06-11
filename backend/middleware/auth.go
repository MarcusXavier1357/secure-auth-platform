package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/repository"
	"auth-backend/internal/service"
)

// Auth valida o JWT RS256, verifica se o usuário segue ativo, atualiza
// last_activity_at da sessão e injeta userId no contexto.
func Auth(auth *service.AuthService, users *repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
		}

		claims, err := auth.ValidateAccessToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		user, err := users.FindByID(c.Context(), claims.UserID)
		if err != nil || !user.Active {
			return fiber.NewError(fiber.StatusUnauthorized, "user not found or inactive")
		}

		auth.TouchSession(c.Context(), claims.SessionID)

		c.Locals("userId", claims.UserID)
		c.Locals("sessionId", claims.SessionID)
		return c.Next()
	}
}

// UserID lê o userId injetado pelo middleware Auth.
func UserID(c *fiber.Ctx) int64 {
	id, _ := c.Locals("userId").(int64)
	return id
}
