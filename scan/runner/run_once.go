package runner

import (
	"context"
	"fmt"

	"github.com/jiajia556/chainkit/core/ports"
	"github.com/jiajia556/chainkit/core/types"
	"github.com/jiajia556/chainkit/scan/cursor"
)

// RunOnce runs a single scan iteration for the given chain + asset.
// It only persists NextCursor after all events are successfully handled.
func (r *Runner) RunOnce(ctx context.Context, chain types.Chain, asset types.Asset) error {
	if r == nil {
		return fmt.Errorf("runner is nil")
	}
	if r.Scanner == nil {
		return fmt.Errorf("runner.Scanner is nil")
	}
	if r.Store == nil {
		return fmt.Errorf("runner.Store is nil")
	}
	if r.Handle == nil {
		return fmt.Errorf("runner.Handle is nil")
	}
	if r.ScannerName == "" {
		return fmt.Errorf("runner.ScannerName is empty")
	}

	limit := r.Limit
	if limit <= 0 {
		limit = 100
	}

	cur, ok, err := r.Store.Get(ctx, chain, r.ScannerName)
	if err != nil {
		return fmt.Errorf("cursor store get: %w", err)
	}
	if !ok || cur == nil {
		cur = cursor.Cursor{}
	}

	req := ports.ScanRequest{
		Chain:         chain,
		Asset:         asset, // allowed to be empty
		Confirmations: r.Confirmations,
		Cursor:        cur,
		Limit:         limit,
	}

	resp, err := r.Scanner.ScanEvents(ctx, req)
	if err != nil {
		return fmt.Errorf("scan deposits: %w", err)
	}

	for _, ev := range resp.Events {
		if err := r.Handle(ctx, ev); err != nil {
			// do NOT advance cursor on partial failure
			return fmt.Errorf("handle event tx=%s idx=%s: %w", ev.TxHash, ev.Index, err)
		}
	}

	if err := r.Store.Set(ctx, chain, r.ScannerName, resp.NextCursor); err != nil {
		return fmt.Errorf("cursor store set: %w", err)
	}
	return nil
}
