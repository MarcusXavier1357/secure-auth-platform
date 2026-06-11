package service

import "testing"

func TestMatchPermission(t *testing.T) {
	granted := []string{"users.*", "audit_logs.read"}

	cases := []struct {
		required string
		want     bool
	}{
		{"users.manage", true},
		{"users.read", true},
		{"permissions.manage", false},
		{"audit_logs.read", true},
	}

	for _, tc := range cases {
		if got := matchPermission(granted, tc.required); got != tc.want {
			t.Errorf("matchPermission(%v, %q) = %v, want %v", granted, tc.required, got, tc.want)
		}
	}

	if !matchPermission([]string{"*"}, "anything.here") {
		t.Error("global wildcard should match any permission")
	}
}
