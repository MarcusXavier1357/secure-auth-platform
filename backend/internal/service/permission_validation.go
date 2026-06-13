package service

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrForbidden              = errors.New("forbidden")
	ErrPermissionCodeTaken    = errors.New("permission code already exists")
	ErrInvalidPermissionCode  = errors.New("invalid permission code")
	ErrProtectedPermission    = errors.New("protected permission")
	ErrPermissionInUse        = errors.New("permission in use")
	ErrLastAdmin              = errors.New("last admin")
)

var permissionCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

func validatePermissionCode(code string) error {
	if code == "*" || code == "users.*" {
		return ErrInvalidPermissionCode
	}
	if strings.HasSuffix(code, ".manage") {
		return ErrInvalidPermissionCode
	}
	if !permissionCodePattern.MatchString(code) {
		return ErrInvalidPermissionCode
	}
	return nil
}

func isProtectedPermissionCode(code string) bool {
	switch code {
	case "*", "users.*", "audit_logs.read":
		return true
	default:
		return false
	}
}
