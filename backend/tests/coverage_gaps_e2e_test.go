package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"auth-backend/internal/app"
)

func testAppConfig() app.Config {
	cfg := baseTestConfig()
	cfg.RedisKeyPrefix = "testgap:"
	return cfg
}

func spawnApp(t *testing.T, cfg app.Config) *app.App {
	t.Helper()
	a, err := app.New(cfg)
	if err != nil {
		t.Fatalf("spawn app: %v", err)
	}
	t.Cleanup(func() { a.Close() })
	return a
}

func TestGetUserByID(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	user := createUser(t, admin, "Por ID", "por-id@test.dev", "Senha12345!")

	resp := admin.do("GET", "/users/"+strconv.FormatInt(user.ID, 10), nil)
	requireStatus(t, resp, http.StatusOK)

	var got userResponse
	decodeJSON(t, resp, &got)
	if got.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, got.Email)
	}

	resp = admin.do("GET", "/users/999999", nil)
	requireStatus(t, resp, http.StatusNotFound)

	resp = admin.do("GET", "/users/abc", nil)
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestGrantRevokePermissionValidation(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	user := createUser(t, admin, "Perm Val", "perm-val@test.dev", "Senha12345!")

	resp := admin.do("POST", "/users/abc/permissions", map[string]int64{"permissionId": 1})
	requireStatus(t, resp, http.StatusBadRequest)

	resp = admin.do("POST", "/users/"+strconv.FormatInt(user.ID, 10)+"/permissions", map[string]string{})
	requireStatus(t, resp, http.StatusBadRequest)

	resp = admin.do("DELETE", "/users/"+strconv.FormatInt(user.ID, 10)+"/permissions/abc", nil)
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestHealthPostgresUnavailable(t *testing.T) {
	a := spawnApp(t, testAppConfig())
	if err := a.DB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := a.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("health request: %v", err)
	}
	requireStatus(t, resp, http.StatusServiceUnavailable)
}

func TestHealthRedisUnavailable(t *testing.T) {
	a := spawnApp(t, testAppConfig())
	if err := a.Redis.Close(); err != nil {
		t.Fatalf("close redis: %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := a.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("health request: %v", err)
	}
	requireStatus(t, resp, http.StatusServiceUnavailable)
}

func TestLoginRedisUnavailableReturns503(t *testing.T) {
	a := spawnApp(t, testAppConfig())
	if err := a.Redis.Close(); err != nil {
		t.Fatalf("close redis: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"email": adminEmail, "password": adminPassword})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	requireStatus(t, resp, http.StatusServiceUnavailable)
}

func TestExpiredAccessTokenRejected(t *testing.T) {
	cfg := testAppConfig()
	cfg.RedisKeyPrefix = "testexp:"
	cfg.AccessTTL = time.Millisecond
	a := spawnApp(t, cfg)

	c := &client{t: t, app: a, cookies: map[string]*http.Cookie{}}
	c.mustLogin(adminEmail, adminPassword)

	time.Sleep(5 * time.Millisecond)

	resp := c.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusUnauthorized)
}
