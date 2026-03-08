package ports

import (
	"context"

	"github.com/jiajia556/chainkit/core/types"
)

type Signer interface {
	Chain() types.Chain
	Address(ctx context.Context) (string, error)

	// Sign and send are chain-impl specific; keep Signer minimal.
	// Concrete chain adapters will accept Signer and construct tx.
}
