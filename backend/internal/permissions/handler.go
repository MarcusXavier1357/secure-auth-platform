package permissions

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/httputil"
	"auth-backend/internal/repository"
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
	userID, err := httputil.ParsePositiveInt64(c.Params("id"), "user id")
	if err != nil {
		return err
	}

	var req grantRequest
	if err := c.BodyParser(&req); err != nil || req.PermissionID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "permissionId is required")
	}

	if err := h.perms.Grant(c.Context(), middleware.UserID(c), userID, req.PermissionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user or permission not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Revoke(c *fiber.Ctx) error {
	userID, err := httputil.ParsePositiveInt64(c.Params("id"), "user id")
	if err != nil {
		return err
	}
	permissionID, err := httputil.ParsePositiveInt64(c.Params("permissionId"), "permissionId")
	if err != nil {
		return err
	}

	if err := h.perms.Revoke(c.Context(), middleware.UserID(c), userID, permissionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user or permission not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
