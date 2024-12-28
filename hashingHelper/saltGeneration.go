package hashingHelper

import (
	"context"
	"crypto/rand"
	"github.com/pkg/errors"
)

var SaltGenerationError = errors.New("Failed to generate salt")

// GenerateSalt generates a random salt with the length specified in the HashParams struct
func (params *HashParams) GenerateSalt(ctx context.Context) ([]byte, error) {
	ctx, span := tracer.Start(ctx, "GenerateSalt")
	defer span.End()

	// Generate a random salt with the specified length
	randomByteSlice := make([]byte, params.SaltLength)
	_, err := rand.Read(randomByteSlice)
	if err != nil {
		err = errors.Wrap(err, SaltGenerationError.Error())
		logger.Error(ctx, err)
		return nil, SaltGenerationError
	}

	return randomByteSlice, nil
}
