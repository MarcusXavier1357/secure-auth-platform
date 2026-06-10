package users

import (
	"errors"
	"net/mail"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/repository"
	"auth-backend/internal/service"
	"auth-backend/middleware"
)

type Handler struct {
	users *service.UserService
	perms *service.PermissionService
}

func NewHandler(users *service.UserService, perms *service.PermissionService) *Handler {
	return &Handler{users: users, perms: perms}
}

func (h *Handler) List(c *fiber.Ctx) error {
	list, err := h.users.List(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(list)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	user, err := h.users.FindByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return err
	}
	return c.JSON(user)
}

// Me retorna o usuário autenticado com seus codes de permissão —
// usado pelo frontend para controle visual.
func (h *Handler) Me(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	user, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return err
	}
	codes, err := h.perms.ListCodesByUser(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"user": user, "permissions": codes})
}

type createRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   *int64 `json:"roleId"`
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req createRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Name == "" || req.Email == "" || len(req.Password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "name, email and password (min 8 chars) are required")
	}
	if !validEmail(req.Email) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid email format")
	}

	user, err := h.users.Create(c.Context(), middleware.UserID(c), service.CreateUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		RoleID:   req.RoleID,
	})
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			return fiber.NewError(fiber.StatusConflict, "email already in use")
		}
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

type updateRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	RoleID   *int64  `json:"roleId"`
	Active   *bool   `json:"active"`
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req updateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Password != nil && len(*req.Password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "password must have at least 8 chars")
	}
	if req.Email != nil && !validEmail(*req.Email) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid email format")
	}

	user, err := h.users.Update(c.Context(), middleware.UserID(c), id, service.UpdateUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		RoleID:   req.RoleID,
		Active:   req.Active,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		if errors.Is(err, service.ErrCannotDeactivateSelf) {
			return fiber.NewError(fiber.StatusConflict, "cannot deactivate your own account")
		}
		return err
	}
	return c.JSON(user)
}

// validEmail exige um endereço simples (sem display name), ex.: user@dominio.com.
func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func parseID(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	return id, nil
}
