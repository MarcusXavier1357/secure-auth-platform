package service

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	if err := validateEmail("user@test.dev"); err != nil {
		t.Errorf("expected valid email: %v", err)
	}
	if err := validateEmail("invalid"); err == nil {
		t.Error("expected invalid email error")
	}
}

func TestValidatePasswordSecure(t *testing.T) {
	tests := []struct {
		name     string
		password string
		userName string
		email    string
		wantErr  error
	}{
		{
			name:     "Valid password",
			password: "MinhaSenha2026",
			wantErr:  nil,
		},
		{
			name:     "Valid password 2",
			password: "Fortaleza123ABC",
			wantErr:  nil,
		},
		{
			name:     "Too short",
			password: "Abc1",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "Too long",
			password: strings.Repeat("a", 129),
			wantErr:  ErrPasswordTooLong,
		},
		{
			name:     "Missing uppercase",
			password: "minhasenha2026",
			wantErr:  ErrPasswordComplexity,
		},
		{
			name:     "Missing lowercase",
			password: "MINHASENHA2026",
			wantErr:  ErrPasswordComplexity,
		},
		{
			name:     "Missing digit",
			password: "MinhaSenhaSemNumero",
			wantErr:  ErrPasswordComplexity,
		},
		{
			name:     "Common password of valid length but weak",
			password: "Senha12345678", // contains "senha123456" which is in commonPasswords (lowercase match)
			wantErr:  ErrPasswordWeak,
		},
		{
			name:     "Ascending sequence",
			password: "Minha12345678",
			wantErr:  ErrPasswordWeak,
		},
		{
			name:     "Descending sequence",
			password: "Minha87654321",
			wantErr:  ErrPasswordWeak,
		},
		{
			name:     "Keyboard sequence",
			password: "Minhaqwertyui1",
			wantErr:  ErrPasswordWeak,
		},
		{
			name:     "Excessive repetitions 1 char",
			password: "A1aaaaaaaaaa", // contains repeating "a" of length 10
			wantErr:  ErrPasswordWeak,
		},
		{
			name:     "Excessive repetitions with complexity chunk 3",
			password: "Ab1Ab1Ab1Ab1Ab1", // repeats "Ab1" 5 times -> length 15 >= 10
			wantErr:  ErrPasswordWeak,
		},
		{
			name:     "Personal data - name",
			password: "MarcusXavier123",
			userName: "Marcus Xavier",
			wantErr:  ErrPasswordWeak,
		},
		{
			name:     "Personal data - email prefix",
			password: "Marcus12345678",
			email:    "marcus.xavier@test.com",
			wantErr:  ErrPasswordWeak,
		},
		{
			name:     "HIBP pwned password",
			password: "Password123456", // highly leaked, has complexity and length
			wantErr:  ErrPasswordPwned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswordSecure(tt.password, tt.userName, tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validatePasswordSecure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
