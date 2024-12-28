package hashingHelper

import (
	"github.com/LucaSchmitz2003/FlowWatch"
	"go.opentelemetry.io/otel"
	"golang.org/x/crypto/argon2"
)

var (
	tracer = otel.Tracer("HashingTracer")
	logger = FlowWatch.GetLogHelper()

	// DefaultHashParams are the default parameters for the hashing algorithm
	DefaultHashParams = HashParams{
		Time:       2,
		Memory:     256 * 1024, // 256 MiB
		Threads:    4,
		KeyLen:     32,
		SaltLength: 32,
		Type:       Argon2id,
		Version:    argon2.Version,
	}
)

// HashParams are the parameters for the hashing algorithm.
// Description according to RFC-9106.
type HashParams struct {
	Time       uint32     // Number of iterations/passes over the memory (-> time) used by the algorithm
	Memory     uint32     // Amount of memory used by the algorithm in KiB (minimum: 8*Threads)
	Threads    uint8      // Parallelism factor determines how many independent (but synchronizing) computational chains (lanes) can be run
	KeyLen     uint32     // Length of the generated hash (key) in bytes. Recommended: 16 or more
	SaltLength uint32     // Length of the random salt in bytes. Recommended: 16 or more
	Type       Argon2Type // Type of the algorithm: 0 = Argon2d, 1 = Argon2i, 2 = Argon2id
	Version    uint32     // Version of the algorithm
}

type Argon2Type uint8

// Argon2 types according to the RFC
const (
	Argon2d Argon2Type = iota
	Argon2i
	Argon2id
)

// String method for Argon2Type
func (t Argon2Type) String() string {
	types := [...]string{"argon2d", "argon2i", "argon2id"}
	if t < Argon2d || t > Argon2id {
		return "unknown"
	}
	return types[t]
}
