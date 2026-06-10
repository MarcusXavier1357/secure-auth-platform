package audit

import (
	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/service"
)

type Handler struct {
	audit *service.AuditService
}

func NewHandler(audit *service.AuditService) *Handler {
	return &Handler{audit: audit}
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	logs, err := h.audit.List(c.Context(), limit, offset)
	if err != nil {
		return err
	}
	return c.JSON(logs)
}
