package ports

import (
	"context"

	"github.com/jiajia556/chainkit/core/types"
)

type TxQueryPort interface {
	GetTxStatus(ctx context.Context, chain types.Chain, txHash string) (types.TxResult, error)
}
