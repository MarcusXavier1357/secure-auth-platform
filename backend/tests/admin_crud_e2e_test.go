package tests

import (
	"fmt"
	"net/http"
	"testing"
)

func grantPermission(t *testing.T, admin *client, userID, permissionID int64) {
	t.Helper()
	resp := admin.do("POST", fmt.Sprintf("/users/%d/permissions", userID),
		map[string]int64{"permissionId": permissionID})
	requireStatus(t, resp, http.StatusNoContent)
}

func TestGranularUserCreateDenied(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	user := createUser(t, admin, "Leitor", "leitor-granular@test.dev", "UnleakedPass2026!")
	readID := findPermissionID(t, admin, "users.read")
	grantPermission(t, admin, user.ID, readID)

	reader := newClient(t)
	reader.mustLogin(user.Email, "UnleakedPass2026!")

	resp := reader.do("GET", "/users", nil)
	requireStatus(t, resp, http.StatusOK)

	resp = reader.do("POST", "/users", map[string]string{
		"name": "Novo", "email": "novo-granular@test.dev", "password": "UnleakedPass2026!",
	})
	requireStatus(t, resp, http.StatusForbidden)
}

func TestGranularPatchFieldPermissions(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	target := createUser(t, admin, "Patch Alvo", "patch-alvo@test.dev", "UnleakedPass2026!")
	editor := createUser(t, admin, "Patch Editor", "patch-editor@test.dev", "UnleakedPass2026!")

	updateID := findPermissionID(t, admin, "users.update")
	grantPermission(t, admin, editor.ID, updateID)

	c := newClient(t)
	c.mustLogin(editor.Email, "UnleakedPass2026!")

	resp := c.do("PATCH", fmt.Sprintf("/users/%d", target.ID),
		map[string]string{"name": "Renomeado"})
	requireStatus(t, resp, http.StatusOK)

	resp = c.do("PATCH", fmt.Sprintf("/users/%d", target.ID),
		map[string]string{"password": "OutraSenha1!"})
	requireStatus(t, resp, http.StatusForbidden)

	resp = c.do("PATCH", fmt.Sprintf("/users/%d", target.ID),
		map[string]bool{"active": false})
	requireStatus(t, resp, http.StatusForbidden)
}

func TestGrantAntiEscalation(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	actor := createUser(t, admin, "Sem Escala", "sem-escala@test.dev", "UnleakedPass2026!")
	grantID := findPermissionID(t, admin, "permissions.grant")
	auditID := findPermissionID(t, admin, "audit_logs.read")

	grantPermission(t, admin, actor.ID, grantID)

	c := newClient(t)
	c.mustLogin(actor.Email, "UnleakedPass2026!")

	// Pode conceder audit_logs.read porque admin já deu essa perm ao actor? No - actor only has permissions.grant, not audit_logs.read
	resp := c.do("POST", fmt.Sprintf("/users/%d/permissions", actor.ID),
		map[string]int64{"permissionId": auditID})
	requireStatus(t, resp, http.StatusForbidden)
}

func TestPermissionCatalogCRUD(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	code := "reports.read"
	resp := admin.do("POST", "/permissions", map[string]string{
		"code": code, "description": "Ler relatórios",
	})
	requireStatus(t, resp, http.StatusCreated)

	var created struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
	}
	decodeJSON(t, resp, &created)
	if created.Code != code {
		t.Fatalf("expected code %s, got %s", code, created.Code)
	}

	resp = admin.do("PATCH", fmt.Sprintf("/permissions/%d", created.ID),
		map[string]string{"description": "Ler relatórios atualizado"})
	requireStatus(t, resp, http.StatusOK)

	resp = admin.do("DELETE", fmt.Sprintf("/permissions/%d", created.ID), nil)
	requireStatus(t, resp, http.StatusNoContent)

	// Código inválido
	resp = admin.do("POST", "/permissions", map[string]string{
		"code": "*", "description": "não permitido",
	})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestProtectedPermissionDelete(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	starID := findPermissionID(t, admin, "*")
	resp := admin.do("DELETE", fmt.Sprintf("/permissions/%d", starID), nil)
	requireStatus(t, resp, http.StatusConflict)
}

func TestPermissionDeleteInUse(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	resp := admin.do("POST", "/permissions", map[string]string{
		"code": "temp.perm", "description": "temporária",
	})
	requireStatus(t, resp, http.StatusCreated)
	var created struct {
		ID int64 `json:"id"`
	}
	decodeJSON(t, resp, &created)

	user := createUser(t, admin, "Com Temp", "com-temp@test.dev", "UnleakedPass2026!")
	grantPermission(t, admin, user.ID, created.ID)

	resp = admin.do("DELETE", fmt.Sprintf("/permissions/%d", created.ID), nil)
	requireStatus(t, resp, http.StatusConflict)
}

func TestLastAdminCannotDeactivate(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	resp := admin.do("GET", "/me", nil)
	requireStatus(t, resp, http.StatusOK)
	var me struct {
		User userResponse `json:"user"`
	}
	decodeJSON(t, resp, &me)

	// Garante que só o admin seed tem * ativo neste banco de teste isolado.
	other := createUser(t, admin, "Outro User", "outro-admin@test.dev", "UnleakedPass2026!")
	resp = admin.do("PATCH", fmt.Sprintf("/users/%d", other.ID), map[string]bool{"active": false})
	requireStatus(t, resp, http.StatusOK)

	resp = admin.do("PATCH", fmt.Sprintf("/users/%d", me.User.ID), map[string]bool{"active": false})
	requireStatus(t, resp, http.StatusConflict)
}

func TestUpdateUserDuplicateEmail(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	createUser(t, admin, "Email A", "email-a@test.dev", "UnleakedPass2026!")
	userB := createUser(t, admin, "Email B", "email-b@test.dev", "UnleakedPass2026!")

	resp := admin.do("PATCH", fmt.Sprintf("/users/%d", userB.ID),
		map[string]string{"email": "email-a@test.dev"})
	requireStatus(t, resp, http.StatusConflict)
}

func TestListRoles(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	resp := admin.do("GET", "/roles", nil)
	requireStatus(t, resp, http.StatusOK)

	var roles []struct {
		Name string `json:"name"`
	}
	decodeJSON(t, resp, &roles)
	if len(roles) == 0 {
		t.Fatal("expected at least one role")
	}
}

func TestPermissionCatalogAuditLog(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	code := "audit.test"
	resp := admin.do("POST", "/permissions", map[string]string{
		"code": code, "description": "teste auditoria",
	})
	requireStatus(t, resp, http.StatusCreated)
	var created struct {
		ID int64 `json:"id"`
	}
	decodeJSON(t, resp, &created)

	resp = admin.do("DELETE", fmt.Sprintf("/permissions/%d", created.ID), nil)
	requireStatus(t, resp, http.StatusNoContent)

	resp = admin.do("GET", "/audit-logs?limit=50", nil)
	requireStatus(t, resp, http.StatusOK)

	var logs []struct {
		Action  string         `json:"action"`
		NewData map[string]any `json:"newData"`
	}
	decodeJSON(t, resp, &logs)

	foundCreate := false
	for _, log := range logs {
		if log.Action == "permission.created" {
			if codeVal, ok := log.NewData["code"].(string); ok && codeVal == code {
				foundCreate = true
				break
			}
		}
	}
	if !foundCreate {
		t.Error("expected permission.created audit log with code")
	}
}

func TestGrantAuditIncludesPermissionCode(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	user := createUser(t, admin, "Audit Grant", "audit-grant@test.dev", "UnleakedPass2026!")
	permID := findPermissionID(t, admin, "users.read")

	resp := admin.do("POST", fmt.Sprintf("/users/%d/permissions", user.ID),
		map[string]int64{"permissionId": permID})
	requireStatus(t, resp, http.StatusNoContent)

	resp = admin.do("GET", "/audit-logs?limit=20", nil)
	requireStatus(t, resp, http.StatusOK)

	var logs []struct {
		Action  string         `json:"action"`
		NewData map[string]any `json:"newData"`
	}
	decodeJSON(t, resp, &logs)

	for _, log := range logs {
		if log.Action == "permission.granted" {
			if code, ok := log.NewData["permissionCode"].(string); ok && code == "users.read" {
				return
			}
		}
	}
	t.Error("expected permission.granted audit with permissionCode users.read")
}
