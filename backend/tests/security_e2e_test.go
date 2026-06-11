package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestRefreshTokenReuseEmitsSecurityAlert(t *testing.T) {
	c := newClient(t)
	c.mustLogin(adminEmail, adminPassword)
	oldCookie := *c.refreshCookie()

	resp := c.do("POST", "/auth/refresh", nil)
	requireStatus(t, resp, http.StatusOK)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&oldCookie)
	reuseResp, err := testApp.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("reuse refresh: %v", err)
	}
	requireStatus(t, reuseResp, http.StatusUnauthorized)

	if !hasAuditAction(t, "security.alert", "refresh_token_reuse") {
		t.Error("expected security.alert with reason refresh_token_reuse")
	}
}

func TestLoginRateLimitEmitsSecurityAlert(t *testing.T) {
	c := newRateLimitClient(t)
	testIP := "198.51.100.77"

	for i := 1; i <= 3; i++ {
		resp := c.loginFromIP(testIP, "brute-alert@test.dev", "senha-errada")
		requireStatus(t, resp, http.StatusUnauthorized)
	}

	resp := c.loginFromIP(testIP, "brute-alert@test.dev", "senha-errada")
	requireStatus(t, resp, http.StatusTooManyRequests)

	if !hasAuditAction(t, "security.alert", "login_rate_limit") {
		t.Error("expected security.alert with reason login_rate_limit")
	}
}

func TestArgon2RehashOnLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}
	_, err = testApp.DB.NewUpdate().
		TableExpr("users").
		Set("password_hash = ?", string(hash)).
		Where("email = ?", adminEmail).
		Exec(context.Background())
	if err != nil {
		t.Fatalf("set bcrypt hash: %v", err)
	}

	c := newClient(t)
	resp := c.login(adminEmail, adminPassword)
	requireStatus(t, resp, http.StatusOK)

	var stored string
	err = testApp.DB.NewRaw("SELECT password_hash FROM users WHERE email = ?", adminEmail).
		Scan(context.Background(), &stored)
	if err != nil {
		t.Fatalf("read password hash: %v", err)
	}
	if len(stored) < 10 || stored[:10] != "$argon2id$" {
		prefix := stored
		if len(prefix) > 20 {
			prefix = prefix[:20]
		}
		t.Errorf("expected argon2id hash after login, got prefix %q", prefix)
	}
}

func TestImpossibleTravelEmitsSecurityAlert(t *testing.T) {
	c := newGeoClient(t)

	resp := c.loginFromIP("203.0.113.1", adminEmail, adminPassword)
	requireStatus(t, resp, http.StatusOK)

	c2 := newGeoClient(t)
	resp = c2.loginFromIP("203.0.113.2", adminEmail, adminPassword)
	requireStatus(t, resp, http.StatusOK)

	if !hasAuditAction(t, "security.alert", "impossible_travel") {
		t.Error("expected security.alert with reason impossible_travel")
	}
}

func hasAuditAction(t *testing.T, action, reason string) bool {
	t.Helper()
	var rows []struct {
		Action  string         `bun:"action"`
		NewData map[string]any `bun:"new_data,type:jsonb"`
	}
	err := testApp.DB.NewSelect().
		TableExpr("audit_logs").
		Column("action", "new_data").
		Where("action = ?", action).
		Order("id DESC").
		Limit(50).
		Scan(context.Background(), &rows)
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	for _, row := range rows {
		if r, _ := row.NewData["reason"].(string); r == reason {
			return true
		}
	}
	return false
}

func (c *client) loginFromIP(ip, email, password string) *http.Response {
	c.t.Helper()

	raw, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		c.t.Fatalf("marshal body: %v", err)
	}

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", ip)
	req.RemoteAddr = ip + ":12345"

	resp, err := c.app.Fiber.Test(req, -1)
	if err != nil {
		c.t.Fatalf("login from %s failed: %v", ip, err)
	}
	for _, ck := range resp.Cookies() {
		c.cookies[ck.Name] = ck
	}
	if resp.StatusCode == http.StatusOK {
		var body struct {
			AccessToken string `json:"accessToken"`
		}
		decodeJSON(c.t, resp, &body)
		c.token = body.AccessToken
	}
	return resp
}
