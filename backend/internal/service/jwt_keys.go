package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type JWTKeyPair struct {
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
}

// LoadJWTKeys carrega par RSA de arquivos PEM.
func LoadJWTKeys(privatePath, publicPath string) (*JWTKeyPair, error) {
	if privatePath == "" || publicPath == "" {
		return nil, errors.New("JWT_PRIVATE_KEY_PATH and JWT_PUBLIC_KEY_PATH are required")
	}
	priv, err := loadPrivateKey(privatePath)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}
	pub, err := loadPublicKey(publicPath)
	if err != nil {
		return nil, fmt.Errorf("load public key: %w", err)
	}
	return &JWTKeyPair{Private: priv, Public: pub}, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid private key PEM")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsaPub, nil
}

// GenerateTestRSAKeyPair gera par RSA para testes e2e.
func GenerateTestRSAKeyPair() (*JWTKeyPair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &JWTKeyPair{Private: priv, Public: &priv.PublicKey}, nil
}
