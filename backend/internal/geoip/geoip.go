package geoip

import "errors"

var ErrUnavailable = errors.New("geoip lookup unavailable")

// Lookup resolve país (ISO 3166-1 alpha-2) a partir de um IP.
type Lookup interface {
	CountryCode(ip string) (string, error)
}

// Nop desabilita geo quando GEOIP_DB_PATH está vazio.
type Nop struct{}

func (Nop) CountryCode(string) (string, error) {
	return "", nil
}

// Mock mapeia IPs fixos — usado nos testes e2e.
type Mock struct {
	Countries map[string]string
}

func (m Mock) CountryCode(ip string) (string, error) {
	if code, ok := m.Countries[ip]; ok {
		return code, nil
	}
	return "", nil
}
