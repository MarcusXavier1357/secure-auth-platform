package service

import (
	"errors"
	"net/mail"
)

const minPasswordLen = 8

var (
	ErrInvalidEmail = errors.New("invalid email")
	ErrWeakPassword = errors.New("weak password")
)

func validateEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLen {
		return ErrWeakPassword
	}
	return nil
}
