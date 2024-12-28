package hashingHelper

import (
	"context"
	"crypto/subtle"
	"github.com/pkg/errors"
)

// VerifyHash compares the provided password with the encoded hash
func VerifyHash(ctx context.Context, encodedHash, password string) (bool, error) {
	ctx, span := tracer.Start(ctx, "VerifyHash")
	defer span.End()

	// Decode the hash
	params, salt, hash, err1 := DecodeHash(ctx, encodedHash)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to decode the hash")
		return false, err1
	}

	// Generate the hash using the provided password and the salt from the encoded hash
	newHash := params.hashPasswordWithSaltWithoutEncoding(ctx, []byte(password), salt)

	// Compare the generated hash with the decoded hash
	verified := subtle.ConstantTimeCompare(newHash, hash) == 1
	return verified, nil
}
