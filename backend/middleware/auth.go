package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/repository"
	"auth-backend/internal/service"
)

// Auth valida o JWT (assinatura + expiração), verifica se o usuário segue
// ativo e injeta o userId no contexto.
func Auth(auth *service.AuthService, users *repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
		}

		userID, err := auth.ValidateAccessToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		user, err := users.FindByID(c.Context(), userID)
		if err != nil || !user.Active {
			return fiber.NewError(fiber.StatusUnauthorized, "user not found or inactive")
		}

		c.Locals("userId", userID)
		return c.Next()
	}
}

// UserID lê o userId injetado pelo middleware Auth.
func UserID(c *fiber.Ctx) int64 {
	id, _ := c.Locals("userId").(int64)
	return id
}
