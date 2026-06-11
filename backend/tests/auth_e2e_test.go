package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	// O seed concede todas as permissões (inclui wildcard *).
	if len(body.Permissions) != 5 {
		t.Errorf("expected 5 permissions for admin, got %d", len(body.Permissions))
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

	// O token novo ainda renova antes de testar reuse do antigo.
	resp = c.do("POST", "/auth/refresh", nil)
	requireStatus(t, resp, http.StatusOK)

	// Reuso do token antigo (já rotacionado) dispara revogação em massa.
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&oldCookie)
	oldResp, err := testApp.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("refresh with old cookie failed: %v", err)
	}
	requireStatus(t, oldResp, http.StatusUnauthorized)
}

func TestRefreshCreatesNewSessionRow(t *testing.T) {
	c := newClient(t)
	c.mustLogin(adminEmail, adminPassword)

	before := countSessions(t)

	resp := c.do("POST", "/auth/refresh", nil)
	requireStatus(t, resp, http.StatusOK)

	after := countSessions(t)
	if after != before+1 {
		t.Errorf("expected session count +1 after refresh, before=%d after=%d", before, after)
	}
}

func TestRefreshTokenReuseRevokesAllSessions(t *testing.T) {
	c := newClient(t)
	c.mustLogin(adminEmail, adminPassword)
	oldCookie := *c.refreshCookie()

	resp := c.do("POST", "/auth/refresh", nil)
	requireStatus(t, resp, http.StatusOK)

	// Reuso do token antigo deve falhar e revogar todas as sessões do usuário.
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&oldCookie)
	reuseResp, err := testApp.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("reuse refresh: %v", err)
	}
	requireStatus(t, reuseResp, http.StatusUnauthorized)

	// Sessão nova também foi revogada pelo reuse detection.
	req2 := httptest.NewRequest("POST", "/auth/refresh", nil)
	req2.AddCookie(c.refreshCookie())
	newResp, err := testApp.Fiber.Test(req2, -1)
	if err != nil {
		t.Fatalf("refresh after reuse: %v", err)
	}
	requireStatus(t, newResp, http.StatusUnauthorized)
}

func TestLoginStoresSessionFingerprint(t *testing.T) {
	c := newClient(t)
	resp := c.doWithHeaders("POST", "/auth/login",
		map[string]string{"email": adminEmail, "password": adminPassword},
		map[string]string{"User-Agent": "TestBrowser/1.0"},
	)
	requireStatus(t, resp, http.StatusOK)

	var body struct {
		AccessToken string `json:"accessToken"`
	}
	decodeJSON(t, resp, &body)
	c.token = body.AccessToken

	var session struct {
		IPAddress      *string    `bun:"ip_address"`
		UserAgent      *string    `bun:"user_agent"`
		LastActivityAt *time.Time `bun:"last_activity_at"`
	}
	err := testApp.DB.NewSelect().
		TableExpr("sessions").
		Column("ip_address", "user_agent", "last_activity_at").
		Where("revoked = ?", false).
		Order("id DESC").
		Limit(1).
		Scan(context.Background(), &session)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if session.UserAgent == nil || *session.UserAgent != "TestBrowser/1.0" {
		t.Errorf("expected user agent stored, got %v", session.UserAgent)
	}
	if session.LastActivityAt == nil {
		t.Error("expected last_activity_at on login")
	}
}

func countSessions(t *testing.T) int {
	t.Helper()
	var n int
	if err := testApp.DB.NewRaw("SELECT COUNT(*)::int FROM sessions").Scan(context.Background(), &n); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	return n
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

	// Tier de teste bloqueia na 4ª tentativa.
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
