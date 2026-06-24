package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"auth-backend/internal/models"
	"auth-backend/internal/service"
	"auth-backend/middleware"
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

func (h *Handler) ListSessions(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	currentSessionID := middleware.SessionID(c)

	sessions, err := h.auth.ListActiveSessions(c.Context(), userID)
	if err != nil {
		return err
	}

	type sessionResponse struct {
		models.Session
		IsCurrent bool `json:"isCurrent"`
	}

	response := make([]sessionResponse, 0, len(sessions))
	for _, s := range sessions {
		response = append(response, sessionResponse{
			Session:   s,
			IsCurrent: s.ID == currentSessionID,
		})
	}
	return c.JSON(response)
}

func (h *Handler) RevokeSession(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	sessionID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid session id")
	}

	currentSessionID := middleware.SessionID(c)
	if sessionID == currentSessionID {
		return fiber.NewError(fiber.StatusConflict, "cannot revoke your current session")
	}

	if err := h.auth.RevokeSession(c.Context(), sessionID, userID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) RevokeAllSessionsExceptCurrent(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	currentSessionID := middleware.SessionID(c)

	if err := h.auth.RevokeOtherSessions(c.Context(), userID, currentSessionID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
