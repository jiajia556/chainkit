package ports

import (
	"context"

	"github.com/jiajia556/chainkit/core/types"
)

type TransferOpts struct {
	Memo string // TRON/BTC等可能用到；EVM一般为空
	// Fee strategy fields can be extended later (maxFeePerGas, priorityFee, etc.)
}

type TransferPort interface {
	TransferNative(ctx context.Context, chain types.Chain, from Signer, to string, amount types.Amount, opts TransferOpts) (txHash string, err error)
	TransferToken(ctx context.Context, asset types.Asset, from Signer, to string, amount types.Amount, opts TransferOpts) (txHash string, err error)
}
