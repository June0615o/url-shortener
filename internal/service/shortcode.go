package service

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/panhao/url-shortener/internal/repository"
	"github.com/panhao/url-shortener/internal/util"
)

const (
	DefaultCodeLength = 7
	MaxRetries        = 10
)

var base62Charset = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

type ShortCodeService struct {
	linkRepo *repository.LinkRepo
}

func NewShortCodeService(linkRepo *repository.LinkRepo) *ShortCodeService {
	return &ShortCodeService{
		linkRepo: linkRepo,
	}
}

// Generate creates a unique random short code using crypto/rand.
// Retries up to MaxRetries times in case of collisions.
func (s *ShortCodeService) Generate(ctx context.Context) (string, error) {
	for i := 0; i < MaxRetries; i++ {
		code, err := randomBase62(DefaultCodeLength)
		if err != nil {
			return "", err
		}

		// Skip reserved words
		if util.IsReservedWord(code) {
			continue
		}

		// Check uniqueness
		exists, err := s.linkRepo.IsShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}

	return "", ErrMaxRetriesExceeded
}

// GenerateWithLength creates a random short code with specified length.
func (s *ShortCodeService) GenerateWithLength(ctx context.Context, length int) (string, error) {
	for i := 0; i < MaxRetries; i++ {
		code, err := randomBase62(length)
		if err != nil {
			return "", err
		}

		if util.IsReservedWord(code) {
			continue
		}

		exists, err := s.linkRepo.IsShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}

	return "", ErrMaxRetriesExceeded
}

func randomBase62(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(62))
		if err != nil {
			return "", err
		}
		b[i] = base62Charset[n.Int64()]
	}
	return string(b), nil
}

var ErrMaxRetriesExceeded = &retryError{}

type retryError struct{}

func (e *retryError) Error() string {
	return "failed to generate unique short code after max retries"
}
