package httputil

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ParsePositiveInt64 valida parâmetro de rota como inteiro positivo.
func ParsePositiveInt64(raw, fieldLabel string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+fieldLabel)
	}
	return id, nil
}
