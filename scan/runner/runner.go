package runner

import (
	"context"
	"time"

	"github.com/jiajia556/chainkit/core/ports"
	"github.com/jiajia556/chainkit/core/types"
	"github.com/jiajia556/chainkit/scan/cursor"
)

type Handler func(ctx context.Context, ev types.Event) error

type Runner struct {
	Scanner ports.ScanPort
	Store   cursor.Store

	ScannerName   string
	Interval      time.Duration
	Limit         int
	Confirmations uint64

	Handle Handler
}
