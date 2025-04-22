package services

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type APIKeyService struct{}

func NewAPIKeyService() *APIKeyService {
	return &APIKeyService{}
}

// CreateNewAPIKey generates a new plain text API key and its bcrypt hash.
// It returns the plain text key (show ONCE to the user) and the hash (store in DB).
func (s *APIKeyService) CreateNewAPIKey() (plainTextKey string, hashedKey string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		err = errors.Wrap(err, "failed to generate random bytes")
		return
	}

	plainTextKey = "vc_" + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainTextKey), bcrypt.DefaultCost)
	if err != nil {
		err = errors.Wrap(err, "failed to generate hashed key")
		plainTextKey = ""
		return
	}
	hashedKey = string(hashedBytes)

	return // Return plainTextKey, hashedKey, nil
}

// IsAPIKeyValid compares a provided plain text key against a stored bcrypt hash.
func (s *APIKeyService) IsAPIKeyValid(storedHashedKey, providedPlainTextKey string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(storedHashedKey), []byte(providedPlainTextKey))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, errors.Wrap(err, "failed to compare api key hash")
}
