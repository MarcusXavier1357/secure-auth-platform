package tests

import (
	"net/http"
	"testing"
)

func TestLegacyPermissionsRemovedFromCatalog(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	resp := admin.do("GET", "/permissions", nil)
	requireStatus(t, resp, http.StatusOK)

	var perms []struct {
		Code string `json:"code"`
	}
	decodeJSON(t, resp, &perms)

	for _, p := range perms {
		if p.Code == "users.manage" || p.Code == "permissions.manage" {
			t.Fatalf("legacy permission %q still in catalog", p.Code)
		}
	}
}

func TestLegacyManageCodeDoesNotAuthorizeRead(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	user := createUser(t, admin, "So Leitura", "so-leitura-legacy@test.dev", "Senha12345!")
	readID := findPermissionID(t, admin, "users.read")
	grantPermission(t, admin, user.ID, readID)

	c := newClient(t)
	c.mustLogin(user.Email, "Senha12345!")

	resp := c.do("GET", "/users", nil)
	requireStatus(t, resp, http.StatusOK)

	resp = c.do("POST", "/users", map[string]string{
		"name": "Novo", "email": "novo-legacy@test.dev", "password": "Senha12345!",
	})
	requireStatus(t, resp, http.StatusForbidden)
}

func TestCannotCreateManageSuffixPermission(t *testing.T) {
	admin := newClient(t)
	admin.mustLogin(adminEmail, adminPassword)

	resp := admin.do("POST", "/permissions", map[string]string{
		"code": "reports.manage", "description": "legado proibido",
	})
	requireStatus(t, resp, http.StatusBadRequest)
}
