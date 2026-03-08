package ports

import (
	"context"

	"github.com/jiajia556/chainkit/core/types"
)

type AddressPort interface {
	ValidateAddress(ctx context.Context, chain types.Chain, address string) error
	NormalizeAddress(ctx context.Context, chain types.Chain, address string) (string, error)
}
