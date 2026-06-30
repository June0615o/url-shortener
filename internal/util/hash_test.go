package util

import "testing"

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("test-password-123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "my-secure-password"
	hash, _ := HashPassword(password)

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should succeed with correct password")
	}
	if CheckPassword("wrong-password", hash) {
		t.Error("CheckPassword should fail with wrong password")
	}
}

func TestSHA256Hash(t *testing.T) {
	h1 := SHA256Hash("hello")
	h2 := SHA256Hash("hello")
	h3 := SHA256Hash("world")

	if h1 != h2 {
		t.Error("SHA256Hash should be deterministic")
	}
	if h1 == h3 {
		t.Error("SHA256Hash should produce different hashes for different inputs")
	}
	if len(h1) != 64 {
		t.Errorf("SHA256Hash length = %d, want 64", len(h1))
	}
}

func TestGenerateAPIKey(t *testing.T) {
	fullKey, keyHash, keyPrefix := GenerateAPIKey(42, "sid")
	if fullKey == "" || keyHash == "" || keyPrefix == "" {
		t.Error("GenerateAPIKey should return non-empty values")
	}
	if len(keyPrefix) != 15 {
		t.Errorf("keyPrefix length = %d, want 15", len(keyPrefix))
	}
}
