package service

import "testing"

func TestValidateEmail(t *testing.T) {
	if err := validateEmail("user@test.dev"); err != nil {
		t.Errorf("expected valid email: %v", err)
	}
	if err := validateEmail("invalid"); err == nil {
		t.Error("expected invalid email error")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := validatePassword("12345678"); err != nil {
		t.Errorf("expected valid password: %v", err)
	}
	if err := validatePassword("short"); err == nil {
		t.Error("expected weak password error")
	}
}
