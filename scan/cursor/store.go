package cursor

import (
	"context"

	"github.com/jiajia556/chainkit/core/types"
)

type Store interface {
	Get(ctx context.Context, chain types.Chain, scannerName string) (Cursor, bool, error)
	Set(ctx context.Context, chain types.Chain, scannerName string, cur Cursor) error
}
