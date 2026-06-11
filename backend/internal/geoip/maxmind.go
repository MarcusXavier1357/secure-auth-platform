package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

// MaxMind consulta um banco GeoLite2 local (.mmdb).
type MaxMind struct {
	db *geoip2.Reader
}

func OpenMaxMind(path string) (*MaxMind, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open geoip database: %w", err)
	}
	return &MaxMind{db: db}, nil
}

func (m *MaxMind) CountryCode(ip string) (string, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", nil
	}
	record, err := m.db.Country(parsed)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if record.Country.IsoCode != "" {
		return record.Country.IsoCode, nil
	}
	return "", nil
}

func (m *MaxMind) Close() error {
	return m.db.Close()
}
