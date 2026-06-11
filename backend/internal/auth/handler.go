package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/service"
)

const refreshCookieName = "refresh_token"

type Handler struct {
	auth         *service.AuthService
	cookieSecure bool
	rateWindow   time.Duration
}

func NewHandler(auth *service.AuthService, cookieSecure bool, rateWindow time.Duration) *Handler {
	return &Handler{auth: auth, cookieSecure: cookieSecure, rateWindow: rateWindow}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil || req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	pair, err := h.auth.Login(c.Context(), req.Email, req.Password, c.IP(), c.Get("User-Agent"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRateLimited):
			c.Set("Retry-After", strconv.Itoa(int(h.rateWindow.Seconds())))
			return fiber.NewError(fiber.StatusTooManyRequests, "too many login attempts")
		case errors.Is(err, service.ErrRateLimitDown):
			return fiber.NewError(fiber.StatusServiceUnavailable, "login temporarily unavailable")
		case errors.Is(err, service.ErrInvalidCredentials):
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		default:
			return err
		}
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	return c.JSON(fiber.Map{"accessToken": pair.AccessToken})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies(refreshCookieName)
	if refreshToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing refresh token")
	}

	pair, err := h.auth.Refresh(c.Context(), refreshToken, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidSession) || errors.Is(err, service.ErrTokenReuse) {
			h.clearRefreshCookie(c)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid session")
		}
		return err
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	return c.JSON(fiber.Map{"accessToken": pair.AccessToken})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	if refreshToken := c.Cookies(refreshCookieName); refreshToken != "" {
		if err := h.auth.Logout(c.Context(), refreshToken); err != nil {
			return err
		}
	}
	h.clearRefreshCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) setRefreshCookie(c *fiber.Ctx, token string, expires time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Expires:  expires,
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     "/",
	})
}

func (h *Handler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     "/",
	})
}
