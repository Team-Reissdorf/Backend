package hashingHelper

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

var (
	InvalidEncodedHashError = errors.New("Invalid encoded hash")
)

// EncodeHash encodes the hash and salt using the provided parameters
func (params *HashParams) EncodeHash(ctx context.Context, hash []byte, salt []byte) string {
	_, span := tracer.Start(ctx, "EncodeHash")
	defer span.End()

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format the encoded hash using the provided parameters.
	encodedHash := fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		params.Type, params.Version, params.Memory, params.Time, params.Threads, b64Salt, b64Hash)

	return encodedHash
}

// DecodeHash decodes the hash, parameters and salt from the encoded hash
func DecodeHash(ctx context.Context, encodedHash string) (HashParams, []byte, []byte, error) {
	ctx, span := tracer.Start(ctx, "DecodeHash")
	defer span.End()

	var params HashParams
	var b64Salt, b64Hash string

	// Split the encoded hash into its parts using the '$' delimiter to extract the parameters
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		logger.Error(ctx, InvalidEncodedHashError)
		return params, nil, nil, InvalidEncodedHashError
	}

	// Check if the hash type is valid
	if parts[1] != Argon2id.String() {
		logger.Error(ctx, UnsupportedHashTypeError)
		return params, nil, nil, UnsupportedHashTypeError
	} else {
		params.Type = Argon2id
	}

	// Parse the parameters from the encoded hash
	n1, err1 := fmt.Sscanf(parts[2], "v=%d", &params.Version)
	if n1 != 1 || err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse the version")
		logger.Error(ctx, err1)
		return params, nil, nil, err1
	}
	n2, err2 := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Time, &params.Threads)
	if n2 != 3 || err2 != nil {
		err2 = errors.Wrap(err2, "Failed to parse the parameters")
		logger.Error(ctx, err2)
		return params, nil, nil, err2
	}
	b64Salt = parts[4]
	b64Hash = parts[5]

	// Base64 decode the salt
	salt, err2 := base64.RawStdEncoding.DecodeString(b64Salt)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to decode the salt")
		logger.Error(ctx, err2)
		return params, nil, nil, err2
	}

	// Base64 decode the hash
	hash, err3 := base64.RawStdEncoding.DecodeString(b64Hash)
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to decode the hash")
		logger.Error(ctx, err3)
		return params, nil, nil, err3
	}

	// Set the length of the salt and hash
	params.SaltLength = uint32(len(salt))
	params.KeyLen = uint32(len(hash))

	return params, salt, hash, nil
}
