package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrPasswordTooShort   = errors.New("password must have at least 12 characters")
	ErrPasswordTooLong    = errors.New("password must have at most 128 characters")
	ErrPasswordComplexity = errors.New("password must contain at least one uppercase, one lowercase, and one number")
	ErrPasswordWeak       = errors.New("password is too weak")
	ErrPasswordPwned      = errors.New("password has been pwned")
)

var hibpClient = &http.Client{
	Timeout: 2 * time.Second,
}

var commonPasswords = map[string]bool{
	"123456":      true,
	"123456789":   true,
	"password":    true,
	"senha123":    true,
	"qwerty":      true,
	"admin":       true,
	"12345678":    true,
	"87654321":    true,
	"senha123456": true,
	"mudar123":    true,
	"welcome":     true,
	"letmein":     true,
	"login123":    true,
}

func validateEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	// Fallback/Legacy validation used internally if needed, but we prefer validatePasswordSecure
	if len(password) < 12 {
		return ErrPasswordTooShort
	}
	return nil
}

func validatePasswordSecure(password string, name string, email string) error {
	runes := []rune(password)
	if len(runes) < 12 {
		return ErrPasswordTooShort
	}
	if len(runes) > 128 {
		return ErrPasswordTooLong
	}

	var (
		hasLower bool
		hasUpper bool
		hasDigit bool
	)
	for _, r := range runes {
		if unicode.IsLower(r) {
			hasLower = true
		} else if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLower || !hasUpper || !hasDigit {
		return ErrPasswordComplexity
	}

	// 1. Common passwords
	pLower := strings.ToLower(password)
	if commonPasswords[pLower] {
		return ErrPasswordWeak
	}

	// 2. Sequential patterns (length >= 8)
	if hasSequence(password) {
		return ErrPasswordWeak
	}

	// 3. Repetitive patterns
	if hasRepetitivePatterns(password) {
		return ErrPasswordWeak
	}

	// 4. Personal data check
	if containsPersonalData(password, name, email) {
		return ErrPasswordWeak
	}

	// 5. HIBP Leak check (Fail-open)
	isPwned, err := checkHIBP(password)
	if err != nil {
		slog.Warn("HIBP API check failed (failing open)", "error", err)
	} else if isPwned {
		return ErrPasswordPwned
	}

	return nil
}

func hasSequence(s string) bool {
	sLower := strings.ToLower(s)
	n := len(sLower)
	if n < 8 {
		return false
	}

	for i := 0; i <= n-8; i++ {
		// Ascending sequence
		isAsc := true
		for j := 0; j < 7; j++ {
			if sLower[i+j+1] != sLower[i+j]+1 {
				isAsc = false
				break
			}
		}
		if isAsc {
			return true
		}

		// Descending sequence
		isDesc := true
		for j := 0; j < 7; j++ {
			if sLower[i+j+1] != sLower[i+j]-1 {
				isDesc = false
				break
			}
		}
		if isDesc {
			return true
		}
	}

	// Keyboard layout rows check (length >= 8)
	rows := []string{"qwertyuiop", "asdfghjkl", "zxcvbnm"}
	for _, row := range rows {
		revRow := reverseString(row)
		for i := 0; i <= n-8; i++ {
			sub := sLower[i : i+8]
			if strings.Contains(row, sub) || strings.Contains(revRow, sub) {
				return true
			}
		}
	}

	return false
}

func hasRepetitivePatterns(s string) bool {
	n := len(s)
	for k := 1; k <= 3; k++ {
		for i := 0; i <= n-k*3; i++ {
			pattern := s[i : i+k]
			repeats := 1
			for j := i + k; j+k <= n; j += k {
				if s[j:j+k] == pattern {
					repeats++
				} else {
					break
				}
			}
			if repeats*k >= 10 {
				return true
			}
		}
	}
	return false
}

func containsPersonalData(password string, name string, email string) bool {
	pLower := strings.ToLower(password)

	if name != "" {
		parts := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
			return r == ' ' || r == '-' || r == '_' || r == '.'
		})
		for _, part := range parts {
			if len(part) >= 3 && strings.Contains(pLower, part) {
				return true
			}
		}
	}

	if email != "" {
		eLower := strings.ToLower(email)
		if strings.Contains(pLower, eLower) {
			return true
		}
		idx := strings.Index(eLower, "@")
		if idx > 0 {
			prefix := eLower[:idx]
			if len(prefix) >= 3 && strings.Contains(pLower, prefix) {
				return true
			}
		}
	}

	return false
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func checkHIBP(password string) (bool, error) {
	h := sha1.New()
	h.Write([]byte(password))
	shaSum := fmt.Sprintf("%X", h.Sum(nil))

	prefix := shaSum[:5]
	suffix := shaSum[5:]

	req, err := http.NewRequest("GET", "https://api.pwnedpasswords.com/range/"+prefix, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "Secure-Auth-Platform-Validator")

	resp, err := hibpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HIBP API returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(bodyBytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			if parts[0] == suffix {
				return true, nil
			}
		}
	}

	return false, nil
}
