package system

import (
	"context"
	"crypto/sha256"
)

type CredentialSupplier func(ctx context.Context) (Credential, error)

// Credential implementations are used to authenticate entities.
type Credential interface {
	Scheme() string
	Value() string
}

// TokenHash creates a SHA-256 hash of the given string.
func TokenHash(token string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(token))

	hashBytes := hasher.Sum(nil)

	return hashBytes
}
