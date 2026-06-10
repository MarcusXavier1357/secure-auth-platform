package permissions

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/service"
	"auth-backend/middleware"
)

type Handler struct {
	perms *service.PermissionService
}

func NewHandler(perms *service.PermissionService) *Handler {
	return &Handler{perms: perms}
}

func (h *Handler) List(c *fiber.Ctx) error {
	list, err := h.perms.ListAll(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(list)
}

type grantRequest struct {
	PermissionID int64 `json:"permissionId"`
}

func (h *Handler) Grant(c *fiber.Ctx) error {
	userID, err := parseUserID(c)
	if err != nil {
		return err
	}

	var req grantRequest
	if err := c.BodyParser(&req); err != nil || req.PermissionID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "permissionId is required")
	}

	if err := h.perms.Grant(c.Context(), middleware.UserID(c), userID, req.PermissionID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Revoke(c *fiber.Ctx) error {
	userID, err := parseUserID(c)
	if err != nil {
		return err
	}
	permissionID, err := strconv.ParseInt(c.Params("permissionId"), 10, 64)
	if err != nil || permissionID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid permissionId")
	}

	if err := h.perms.Revoke(c.Context(), middleware.UserID(c), userID, permissionID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseUserID(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}
	return id, nil
}
