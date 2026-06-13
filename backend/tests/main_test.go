// Package tests contém os testes ponta a ponta da API.
//
// Pré-requisito: Postgres e Redis acessíveis (docker compose up -d postgres redis).
// A suite cria um banco isolado `auth_test` e usa o Redis db 1, sem tocar nos
// dados de desenvolvimento.
package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun/driver/pgdriver"

	"auth-backend/cache"
	"auth-backend/internal/app"
	"auth-backend/internal/geoip"
	"auth-backend/internal/password"
	"auth-backend/internal/service"
)

const (
	adminEmail    = "admin@test.dev"
	adminPassword = "AdminTest123!"
)

var (
	testApp      *app.App
	rateLimitApp *app.App
	geoTestApp   *app.App
	testJWTKeys  *service.JWTKeyPair
)

func TestMain(m *testing.M) {
	adminDSN := getenv("TEST_PG_ADMIN_URL",
		"postgres://auth:auth_dev_password@127.0.0.1:55432/auth?sslmode=disable")
	redisURL := getenv("TEST_REDIS_URL", "redis://127.0.0.1:6379/1")

	if err := recreateTestDatabase(adminDSN); err != nil {
		log.Fatalf("e2e tests require postgres+redis (run: docker compose up -d postgres redis): %v", err)
	}
	if err := flushTestRedis(redisURL); err != nil {
		log.Fatalf("e2e tests require redis: %v", err)
	}

	password.SetParams(password.TestParams())

	var err error
	testJWTKeys, err = service.GenerateTestRSAKeyPair()
	if err != nil {
		log.Fatalf("generate test jwt keys: %v", err)
	}

	baseCfg := baseTestConfig()
	baseCfg.RedisKeyPrefix = "test:"

	testApp, err = app.New(baseCfg)
	if err != nil {
		log.Fatalf("failed to build test app: %v", err)
	}
	if err := testApp.SeedAdmin(context.Background()); err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	rlCfg := baseCfg
	rlCfg.RedisKeyPrefix = "testrl:"
	rlCfg.LoginRateTiers = []cache.RateTier{{Threshold: 4, Block: 15 * time.Minute}}
	rateLimitApp, err = app.New(rlCfg)
	if err != nil {
		log.Fatalf("failed to build rate limit app: %v", err)
	}

	geoCfg := baseCfg
	geoCfg.RedisKeyPrefix = "testgeo:"
	geoCfg.GeoIPLookup = geoip.Mock{
		Countries: map[string]string{
			"203.0.113.1": "BR",
			"203.0.113.2": "US",
		},
	}
	geoCfg.ImpossibleTravelWindow = 30 * time.Minute
	geoTestApp, err = app.New(geoCfg)
	if err != nil {
		log.Fatalf("failed to build geo test app: %v", err)
	}
	if err := geoTestApp.SeedAdmin(context.Background()); err != nil {
		log.Fatalf("failed to seed geo test admin: %v", err)
	}

	code := m.Run()

	testApp.Close()
	rateLimitApp.Close()
	geoTestApp.Close()
	os.Exit(code)
}

func recreateTestDatabase(adminDSN string) error {
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(adminDSN)))
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	if _, err := db.Exec("DROP DATABASE IF EXISTS auth_test WITH (FORCE)"); err != nil {
		return fmt.Errorf("drop test database: %w", err)
	}
	if _, err := db.Exec("CREATE DATABASE auth_test"); err != nil {
		return fmt.Errorf("create test database: %w", err)
	}
	return nil
}

func flushTestRedis(url string) error {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return err
	}
	rdb := redis.NewClient(opts)
	defer func() { _ = rdb.Close() }()
	return rdb.FlushDB(context.Background()).Err()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func baseTestConfig() app.Config {
	return app.Config{
		DatabaseURL:     getenv("TEST_PG_URL", "postgres://auth:auth_dev_password@127.0.0.1:55432/auth_test?sslmode=disable"),
		MigrationsPath:  "../migrations",
		RedisURL:        getenv("TEST_REDIS_URL", "redis://127.0.0.1:6379/1"),
		RedisKeyPrefix:  "test:",
		JWTKeyPair:      testJWTKeys,
		AccessTTL:       15 * time.Minute,
		RefreshTTL:      30 * 24 * time.Hour,
		PermCacheTTL:    5 * time.Minute,
		LoginRateTiers:  app.DefaultLoginRateTiers(),
		LoginCounterTTL: 24 * time.Hour,
		AdminEmail:      adminEmail,
		AdminPassword:   adminPassword,
		CookieSecure:    false,
	}
}

type client struct {
	t       *testing.T
	app     *app.App
	token   string
	cookies map[string]*http.Cookie
}

func newClient(t *testing.T) *client {
	return &client{t: t, app: testApp, cookies: map[string]*http.Cookie{}}
}

func newRateLimitClient(t *testing.T) *client {
	return &client{t: t, app: rateLimitApp, cookies: map[string]*http.Cookie{}}
}

func newGeoClient(t *testing.T) *client {
	return &client{t: t, app: geoTestApp, cookies: map[string]*http.Cookie{}}
}

func (c *client) do(method, path string, body any) *http.Response {
	return c.doWithHeaders(method, path, body, nil)
}

func (c *client) doWithHeaders(method, path string, body any, extraHeaders map[string]string) *http.Response {
	c.t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	if ip := extraHeaders["X-Forwarded-For"]; ip != "" {
		req.RemoteAddr = ip + ":12345"
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}

	resp, err := c.app.Fiber.Test(req, -1)
	if err != nil {
		c.t.Fatalf("%s %s failed: %v", method, path, err)
	}

	for _, ck := range resp.Cookies() {
		c.cookies[ck.Name] = ck
	}
	return resp
}

func (c *client) login(email, password string) *http.Response {
	return c.loginWithIP(email, password, "")
}

func (c *client) loginWithIP(email, password, ip string) *http.Response {
	c.t.Helper()
	resp := c.doWithHeaders("POST", "/auth/login",
		map[string]string{"email": email, "password": password},
		loginIPHeaders(ip),
	)
	if resp.StatusCode == http.StatusOK {
		var body struct {
			AccessToken string `json:"accessToken"`
		}
		decodeJSON(c.t, resp, &body)
		c.token = body.AccessToken
	}
	return resp
}

func loginIPHeaders(ip string) map[string]string {
	if ip == "" {
		return nil
	}
	return map[string]string{"X-Forwarded-For": ip}
}

func (c *client) mustLogin(email, password string) {
	c.t.Helper()
	resp := c.login(email, password)
	if resp.StatusCode != http.StatusOK {
		c.t.Fatalf("login as %s failed with status %d", email, resp.StatusCode)
	}
}

func (c *client) refreshCookie() *http.Cookie {
	return c.cookies["refresh_token"]
}

func decodeJSON(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func requireStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("expected status %d, got %d (body: %s)", want, resp.StatusCode, raw)
	}
}
