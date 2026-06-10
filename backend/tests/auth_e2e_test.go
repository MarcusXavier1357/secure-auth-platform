package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	c := newClient(t)
	resp := c.do("GET", "/health", nil)
	requireStatus(t, resp, http.StatusOK)
}

func TestLoginSuccess(t *testing.T) {
	c := newClient(t)
	resp := c.login(adminEmail, adminPassword)
	requireStatus(t, resp, http.StatusOK)

	if c.token == "" {
		t.Fatal("expected access token in response")
	}

	cookie := c.refreshCookie()
	if cookie == nil {
		t.Fatal("expected refresh_token cookie")
	}
	if !cookie.HttpOnly {
		t.Error("refresh cookie must be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("refresh cookie must be SameSite=Strict, got %v", cookie.SameSite)
	}
	if cookie.Value == "" {
		t.Error("refresh cookie must have a value")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	c := newClient(t)
	resp := c.login(adminEmail, "senha-errada")
	requireStatus(t, resp, http.StatusUnauthorized)
}

func TestLoginUnknownEmail(t *testing.T) {
	c := newClient(t)
	resp := c.login("nao-existe@test.dev", "qualquer-senha")
	requireStatus(t, resp, http.StatusUnauthorized)
}

func TestLoginMissingFields(t *testing.T) {
	c := newClient(t)
	resp := c.do("POST", "/auth/login", map[string]string{"email": adminEmail})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestProtectedRouteWithoutToken(t *testing.T) {
	c := newClient(t)
	resp := c.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusUnauthorized)
}

func TestProtectedRouteWithTamperedToken(t *testing.T) {
	c := newClient(t)
	c.mustLogin(adminEmail, adminPassword)

	// Corrompe a assinatura mantendo o formato JWT.
	parts := strings.Split(c.token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected JWT format")
	}
	c.token = parts[0] + "." + parts[1] + "." + "assinatura-invalida"

	resp := c.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusUnauthorized)
}

func TestMeReturnsUserAndPermissions(t *testing.T) {
	c := newClient(t)
	c.mustLogin(adminEmail, adminPassword)

	resp := c.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusOK)

	var body struct {
		User struct {
			Email string `json:"email"`
		} `json:"user"`
		Permissions []string `json:"permissions"`
	}
	decodeJSON(t, resp, &body)

	if body.User.Email != adminEmail {
		t.Errorf("expected email %s, got %s", adminEmail, body.User.Email)
	}
	// O seed concede as 3 permissões ativas ao admin.
	if len(body.Permissions) != 3 {
		t.Errorf("expected 3 permissions for admin, got %d", len(body.Permissions))
	}
}

func TestRefreshRotatesToken(t *testing.T) {
	c := newClient(t)
	c.mustLogin(adminEmail, adminPassword)

	oldCookie := *c.refreshCookie()

	resp := c.do("POST", "/auth/refresh", nil)
	requireStatus(t, resp, http.StatusOK)

	var body struct {
		AccessToken string `json:"accessToken"`
	}
	decodeJSON(t, resp, &body)
	if body.AccessToken == "" {
		t.Fatal("expected new access token after refresh")
	}

	newCookie := c.refreshCookie()
	if newCookie.Value == oldCookie.Value {
		t.Fatal("refresh token must rotate — same value returned")
	}

	// O token antigo deve ter sido invalidado pela rotação.
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&oldCookie)
	oldResp, err := testApp.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("refresh with old cookie failed: %v", err)
	}
	requireStatus(t, oldResp, http.StatusUnauthorized)

	// O token novo continua válido.
	resp = c.do("POST", "/auth/refresh", nil)
	requireStatus(t, resp, http.StatusOK)
}

func TestRefreshWithoutCookie(t *testing.T) {
	c := newClient(t)
	resp := c.do("POST", "/auth/refresh", nil)
	requireStatus(t, resp, http.StatusUnauthorized)
}

func TestLogoutRevokesSession(t *testing.T) {
	c := newClient(t)
	c.mustLogin(adminEmail, adminPassword)
	sessionCookie := *c.refreshCookie()

	resp := c.do("POST", "/auth/logout", nil)
	requireStatus(t, resp, http.StatusNoContent)

	// A sessão revogada não pode mais ser renovada, mesmo com o cookie antigo.
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&sessionCookie)
	refreshResp, err := testApp.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("refresh after logout failed: %v", err)
	}
	requireStatus(t, refreshResp, http.StatusUnauthorized)
}

func TestLoginRateLimit(t *testing.T) {
	c := newRateLimitClient(t)

	// Limite da app de teste é 3 por janela (IP e email).
	for i := 1; i <= 3; i++ {
		resp := c.login("brute@test.dev", "senha-errada")
		requireStatus(t, resp, http.StatusUnauthorized)
	}

	resp := c.login("brute@test.dev", "senha-errada")
	requireStatus(t, resp, http.StatusTooManyRequests)
	if resp.Header.Get("Retry-After") == "" {
		t.Error("429 response must include Retry-After header")
	}

	// Mesmo a senha correta é bloqueada enquanto a janela durar
	// (rate limit roda antes do bcrypt).
	resp = c.login(adminEmail, adminPassword)
	requireStatus(t, resp, http.StatusTooManyRequests)
}
