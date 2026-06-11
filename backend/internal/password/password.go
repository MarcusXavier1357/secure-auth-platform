package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Params define o custo do Argon2id (OWASP: m=65536, t=3, p=2 para servidores).
type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultParams = Params{
	Memory:      65536,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// TestParams reduz memória/iterações para a suite e2e.
func TestParams() Params {
	return Params{
		Memory:      4096,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

var activeParams = DefaultParams

// SetParams permite ajustar custo (testes ou env ARGON2_MEMORY via app).
func SetParams(p Params) {
	activeParams = p
}

// Hash gera PHC Argon2id: $argon2id$v=19$m=...,t=...,p=...$salt$hash
func Hash(plain string) (string, error) {
	p := activeParams
	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(plain), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Iterations, p.Parallelism, b64Salt, b64Hash), nil
}

// Verify valida bcrypt legado ou Argon2id. needsRehash=true quando bcrypt OK (migração no login).
func Verify(stored, plain string) (needsRehash bool, err error) {
	switch {
	case strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$"):
		if err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)); err != nil {
			return false, err
		}
		return true, nil
	case strings.HasPrefix(stored, "$argon2id$"):
		ok, err := verifyArgon2id(stored, plain)
		return false, errIfNotOK(ok, err)
	default:
		return false, errors.New("unknown password hash format")
	}
}

func errIfNotOK(ok bool, err error) error {
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("password mismatch")
	}
	return nil
}

func verifyArgon2id(stored, plain string) (bool, error) {
	p, salt, hash, err := decodeArgon2id(stored)
	if err != nil {
		return false, err
	}
	other := argon2.IDKey([]byte(plain), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)
	if subtle.ConstantTimeCompare(hash, other) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeArgon2id(stored string) (Params, []byte, []byte, error) {
	parts := strings.Split(stored, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Params{}, nil, nil, errors.New("invalid argon2id hash")
	}
	var version int
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return Params{}, nil, nil, err
	}
	if version != argon2.Version {
		return Params{}, nil, nil, errors.New("unsupported argon2 version")
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return Params{}, nil, nil, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Params{}, nil, nil, err
	}
	return Params{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		KeyLength:   uint32(len(hash)),
	}, salt, hash, nil
}

var dummyHash string

func init() {
	prev := activeParams
	activeParams = TestParams()
	h, err := Hash("timing-equalizer-dummy")
	activeParams = prev
	if err != nil {
		panic(fmt.Sprintf("generating dummy argon2id hash: %v", err))
	}
	dummyHash = h
}

// DummyVerify executa verify contra hash fixo para equalizar timing em falhas de login.
func DummyVerify(plain string) {
	_, _ = Verify(dummyHash, plain)
}
