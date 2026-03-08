package ports

import (
	"context"

	"github.com/jiajia556/chainkit/core/types"
)

type BalancePort interface {
	GetNativeBalance(ctx context.Context, chain types.Chain, address string) (types.Amount, error)
	GetTokenBalance(ctx context.Context, asset types.Asset, address string) (types.Amount, error) // if token unsupported -> ErrUnsupported
}
