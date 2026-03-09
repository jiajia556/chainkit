package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jiajia556/chainkit/adapters/evm"
	"github.com/jiajia556/chainkit/core/ports"
	memorycursor "github.com/jiajia556/chainkit/scan/cursor/memory"
	"github.com/jiajia556/chainkit/scan/runner"
)

// dummyScanner is a demo EVM scanner that returns no events and advances
// the cursor by one block on each call.
type dummyScanner struct{}

func (d *dummyScanner) ScanEvents(_ context.Context, req evm.ScanRequest) (evm.ScanResponse, error) {
	next := evm.Cursor{
		BlockNumber: req.Cursor.BlockNumber + 1,
		LogIndex:    0,
	}
	return evm.NewScanResponse(nil, next), nil
}

// Compile-time assertion: dummyScanner satisfies ports.ScanPort.
var _ ports.ScanPort[evm.ScanRequest, evm.ScanResponse] = (*dummyScanner)(nil)

func main() {
	ctx := context.Background()

	store := memorycursor.NewStore[evm.Cursor]()
	scanner := &dummyScanner{}

	r := &runner.Runner[evm.ScanRequest, evm.Event, evm.Cursor, evm.ScanResponse]{
		Scanner: scanner,
		Store:   store,
		BuildRequest: func(cur evm.Cursor, limit int) evm.ScanRequest {
			return evm.ScanRequest{
				Target: evm.Target{
					ContractAddress: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
					Topic0:          "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				},
				Cursor:        cur,
				Confirmations: 12,
				Limit:         limit,
			}
		},
		Handle: func(_ context.Context, ev evm.Event) error {
			fmt.Printf("event: tx=%s block=%d logIndex=%d\n", ev.TxHash, ev.BlockNumber, ev.LogIndex)
			return nil
		},
		Limit:      100,
		InitCursor: evm.Cursor{BlockNumber: 0},
	}

	if err := r.RunOnce(ctx); err != nil {
		log.Fatalf("RunOnce: %v", err)
	}

	cur, ok, err := store.Get(ctx)
	if err != nil {
		log.Fatalf("store.Get: %v", err)
	}
	fmt.Printf("after RunOnce: ok=%v cursor=%+v\n", ok, cur)
}
