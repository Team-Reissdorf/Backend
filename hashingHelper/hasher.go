package hashingHelper

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

var (
	UnsupportedHashTypeError    = errors.New("Unsupported hash type")
	UnsupportedHashVersionError = errors.New("Unsupported hash version")
)

// HashPassword generates a salt and hashes the provided password using the Argon2id algorithm
func (params *HashParams) HashPassword(ctx context.Context, password string) (string, error) {
	ctx, span := tracer.Start(ctx, "HashPassword")
	defer span.End()

	// Check if the hash type is valid
	if params.Type != Argon2id {
		logger.Error(ctx, UnsupportedHashTypeError)
		return "", UnsupportedHashTypeError
	}

	// Check if the hash version is compatible with the current version of the Argon2id algorithm
	if params.Version != argon2.Version {
		logger.Error(ctx, UnsupportedHashVersionError)
		return "", UnsupportedHashVersionError
	}

	// Generate a random salt using the provided parameters
	salt, err := params.GenerateSalt(ctx)
	if err != nil {
		return "", err
	}

	// Generate the hash
	hash := params.hashPasswordWithSaltWithoutEncoding(ctx, []byte(password), salt)

	// Encode the hash
	encodedHash := params.EncodeHash(ctx, hash, salt)

	return encodedHash, nil
}

// hashPasswordWithSaltWithoutEncoding generates the hash with the given salt
func (params *HashParams) hashPasswordWithSaltWithoutEncoding(ctx context.Context, password, salt []byte) []byte {
	_, span := tracer.Start(ctx, "GenerateHash")
	defer span.End()

	// Generate the hash (key) using the Argon2id algorithm with the provided parameters and the salt
	hash := argon2.IDKey(password, salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	return hash
}
