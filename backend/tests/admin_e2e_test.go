package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type userResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

func createUser(t *testing.T, admin *client, name, email, password string) userResponse {
	t.Helper()
	resp := admin.do("POST", "/users", map[string]string{
		"name": name, "email": email, "password": password,
	})
	requireStatus(t, resp, http.StatusCreated)

	var user userResponse
	decodeJSON(t, resp, &user)
	return user
}

func findPermissionID(t *testing.T, admin *client, code string) int64 {
	t.Helper()
	resp := admin.do("GET", "/permissions", nil)
	requireStatus(t, resp, http.StatusOK)

	var perms []struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
	}
	decodeJSON(t, resp, &perms)

	for _, p := range perms {
		if p.Code == code {
			return p.ID
		}
	}
	t.Fatalf("permission %s not found", code)
	return 0
}

func TestUserManagementRequiresPermission(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	user := createUser(t, admin, "Sem Permissão", "sem-permissao@test.dev", "Senha12345!")

	plain := newClient(t)
	plain.mustLogin(user.Email, "Senha12345!")

	for _, route := range []struct{ method, path string }{
		{"GET", "/users"},
		{"POST", "/users"},
		{"GET", "/permissions"},
		{"GET", "/audit-logs"},
	} {
		resp := plain.do(route.method, route.path, map[string]string{})
		requireStatus(t, resp, http.StatusForbidden)
	}
}

func TestCreateUserValidations(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	// Email duplicado
	createUser(t, admin, "Original", "duplicado@test.dev", "Senha12345!")
	resp := admin.do("POST", "/users", map[string]string{
		"name": "Cópia", "email": "duplicado@test.dev", "password": "Senha12345!",
	})
	requireStatus(t, resp, http.StatusConflict)

	// Senha curta
	resp = admin.do("POST", "/users", map[string]string{
		"name": "Curto", "email": "curto@test.dev", "password": "curta",
	})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestGrantAndRevokePermissionInvalidatesCache(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	user := createUser(t, admin, "Promovido", "promovido@test.dev", "Senha12345!")
	permID := findPermissionID(t, admin, "users.manage")

	promoted := newClient(t)
	promoted.mustLogin(user.Email, "Senha12345!")

	// Sem permissão (e popula o cache com o estado atual)
	resp := promoted.do("GET", "/users", nil)
	requireStatus(t, resp, http.StatusForbidden)

	// Grant deve invalidar o cache imediatamente
	resp = admin.do("POST", fmt.Sprintf("/users/%d/permissions", user.ID),
		map[string]int64{"permissionId": permID})
	requireStatus(t, resp, http.StatusNoContent)

	resp = promoted.do("GET", "/users", nil)
	requireStatus(t, resp, http.StatusOK)

	// Revoke idem
	resp = admin.do("DELETE", fmt.Sprintf("/users/%d/permissions/%d", user.ID, permID), nil)
	requireStatus(t, resp, http.StatusNoContent)

	resp = promoted.do("GET", "/users", nil)
	requireStatus(t, resp, http.StatusForbidden)
}

func TestDeactivatedUserLosesAllAccess(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	user := createUser(t, admin, "Desativado", "desativado@test.dev", "Senha12345!")

	victim := newClient(t)
	victim.mustLogin(user.Email, "Senha12345!")
	sessionCookie := *victim.refreshCookie()

	// Sanidade: acesso funciona antes da desativação
	resp := victim.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusOK)

	resp = admin.do("PATCH", fmt.Sprintf("/users/%d", user.ID),
		map[string]bool{"active": false})
	requireStatus(t, resp, http.StatusOK)

	// Access token existente para de funcionar (middleware verifica ativo)
	resp = victim.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusUnauthorized)

	// Sessões revogadas: refresh falha
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&sessionCookie)
	refreshResp, err := testApp.Fiber.Test(req, -1)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	requireStatus(t, refreshResp, http.StatusUnauthorized)

	// Novo login bloqueado
	relogin := newClient(t)
	resp = relogin.login(user.Email, "Senha12345!")
	requireStatus(t, resp, http.StatusUnauthorized)

	// Reativação restaura o acesso
	resp = admin.do("PATCH", fmt.Sprintf("/users/%d", user.ID),
		map[string]bool{"active": true})
	requireStatus(t, resp, http.StatusOK)

	resp = relogin.login(user.Email, "Senha12345!")
	requireStatus(t, resp, http.StatusOK)
}

func TestAdminCannotDeactivateSelf(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	resp := admin.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusOK)
	var me struct {
		User userResponse `json:"user"`
	}
	decodeJSON(t, resp, &me)

	resp = admin.do("PATCH", fmt.Sprintf("/users/%d", me.User.ID),
		map[string]bool{"active": false})
	requireStatus(t, resp, http.StatusConflict)

	// O acesso continua funcionando — nada foi desativado.
	resp = admin.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusOK)
}

func TestCreateUserInvalidEmail(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	for _, email := range []string{"sem-arroba", "a@b@c", "Nome <nome@test.dev>"} {
		resp := admin.do("POST", "/users", map[string]string{
			"name": "Email Inválido", "email": email, "password": "Senha12345!",
		})
		requireStatus(t, resp, http.StatusBadRequest)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	resp := admin.do("PATCH", "/users/999999", map[string]string{"name": "Fantasma"})
	requireStatus(t, resp, http.StatusNotFound)

	resp = admin.do("PATCH", "/users/abc", map[string]string{"name": "Inválido"})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestAuditLogsRecordCriticalActions(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	// Gera todas as ações críticas dentro do próprio teste para não depender
	// da ordem de execução dos demais.
	user := createUser(t, admin, "Auditado", "auditado@test.dev", "Senha12345!")
	permID := findPermissionID(t, admin, "clients.read")

	resp := admin.do("POST", fmt.Sprintf("/users/%d/permissions", user.ID),
		map[string]int64{"permissionId": permID})
	requireStatus(t, resp, http.StatusNoContent)
	resp = admin.do("DELETE", fmt.Sprintf("/users/%d/permissions/%d", user.ID, permID), nil)
	requireStatus(t, resp, http.StatusNoContent)

	resp = admin.do("PATCH", fmt.Sprintf("/users/%d", user.ID),
		map[string]string{"name": "Auditado Renomeado"})
	requireStatus(t, resp, http.StatusOK)

	audited := newClient(t)
	resp = audited.login(user.Email, "senha-errada")
	requireStatus(t, resp, http.StatusUnauthorized)
	audited.mustLogin(user.Email, "Senha12345!")
	resp = audited.do("POST", "/auth/logout", nil)
	requireStatus(t, resp, http.StatusNoContent)

	resp = admin.do("GET", "/audit-logs?limit=200", nil)
	requireStatus(t, resp, http.StatusOK)

	var logs []struct {
		Action string `json:"action"`
	}
	decodeJSON(t, resp, &logs)

	expected := map[string]bool{
		"login.success":      false,
		"login.failed":       false,
		"logout":             false,
		"user.created":       false,
		"user.updated":       false,
		"permission.granted": false,
		"permission.revoked": false,
	}
	for _, entry := range logs {
		if _, ok := expected[entry.Action]; ok {
			expected[entry.Action] = true
		}
	}
	for action, seen := range expected {
		if !seen {
			t.Errorf("expected audit log action %q to be recorded", action)
		}
	}
}
