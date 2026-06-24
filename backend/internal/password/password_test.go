package password

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndVerifyArgon2id(t *testing.T) {
	SetParams(TestParams())
	t.Cleanup(func() { SetParams(DefaultParams) })

	hash, err := Hash("Senha12345!")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		prefix := hash
		if len(prefix) > 12 {
			prefix = prefix[:12]
		}
		t.Fatalf("expected argon2id prefix, got %q", prefix)
	}

	needsRehash, err := Verify(hash, "Senha12345!")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if needsRehash {
		t.Error("argon2id should not need rehash")
	}

	_, err = Verify(hash, "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestVerifyBcryptNeedsRehash(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("legacy"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	needsRehash, err := Verify(string(hash), "legacy")
	if err != nil {
		t.Fatalf("verify bcrypt: %v", err)
	}
	if !needsRehash {
		t.Error("bcrypt should request rehash")
	}
}

func TestVerifyUnknownFormat(t *testing.T) {
	_, err := Verify("not-a-hash", "plain")
	if err == nil {
		t.Fatal("expected error for unknown hash format")
	}
}

func TestDummyVerifyDoesNotPanic(_ *testing.T) {
	DummyVerify("any-password")
}
