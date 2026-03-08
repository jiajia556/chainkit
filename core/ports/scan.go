package ports

import (
	"context"
	"time"

	"github.com/jiajia556/chainkit/core/types"
	"github.com/jiajia556/chainkit/scan/cursor"
)

type ScanRequest struct {
	Chain types.Chain
	Asset types.Asset // native or token; if empty -> chain-defined default behavior

	// Confirmation depth: scanner should not emit events newer than (latest - Confirmations)
	Confirmations uint64

	// Optional: time range scanning (TRON often easier by timestamp)
	FromTime *time.Time
	ToTime   *time.Time

	// Cursor for incremental scanning
	Cursor cursor.Cursor

	Limit int
}

type ScanResponse struct {
	Events     []types.Event
	NextCursor cursor.Cursor
}

type ScanPort interface {
	ScanDeposits(ctx context.Context, req ScanRequest) (ScanResponse, error)
}
