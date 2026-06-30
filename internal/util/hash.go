package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func SHA256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func GenerateAPIKey(userID int64, prefix string) (fullKey string, keyHash string, keyPrefix string) {
	randomPart := fmt.Sprintf("%d-%s", userID, prefix)
	hash := SHA256Hash(randomPart)
	fullKey = prefix + "_" + hash[:32]
	keyHash = SHA256Hash(fullKey)
	keyPrefix = fullKey[:12] + "..."
	return
}
