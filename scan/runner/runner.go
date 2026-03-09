package runner

import (
	"context"
	"fmt"

	"github.com/jiajia556/chainkit/core/ports"
	"github.com/jiajia556/chainkit/scan/cursor"
)

// Response is the constraint that scan responses must satisfy.
// Ev is the chain-specific event type; Cur is the chain-specific cursor type.
type Response[Ev any, Cur any] interface {
	Events() []Ev
	NextCursor() Cur
}

// Runner orchestrates a single scan iteration using generic types.
//
//   - Req  – chain-specific scan request
//   - Ev   – chain-specific event
//   - Cur  – chain-specific cursor (strongly typed struct)
//   - Resp – chain-specific scan response satisfying Response[Ev, Cur]
type Runner[Req any, Ev any, Cur any, Resp Response[Ev, Cur]] struct {
	// Scanner is the chain-specific ScanPort implementation.
	Scanner ports.ScanPort[Req, Resp]

	// Store persists the cursor between runs.
	Store cursor.Store[Cur]

	// BuildRequest constructs a scan request from the current cursor and limit.
	BuildRequest func(cur Cur, limit int) Req

	// Handle is called for each event. If it returns an error, the cursor is
	// not advanced and RunOnce returns immediately.
	Handle func(ctx context.Context, ev Ev) error

	// Limit is the maximum number of events to fetch per run (default 100).
	Limit int

	// InitCursor is used when the store has no cursor yet.
	InitCursor Cur
}

// RunOnce executes one scan iteration:
//  1. Load cursor from store (or use InitCursor if not set).
//  2. Call Scanner.ScanEvents.
//  3. Call Handle for each event; stop and return on first error.
//  4. Persist NextCursor only after all events are handled successfully.
func (r *Runner[Req, Ev, Cur, Resp]) RunOnce(ctx context.Context) error {
	limit := r.Limit
	if limit <= 0 {
		limit = 100
	}

	cur, ok, err := r.Store.Get(ctx)
	if err != nil {
		return fmt.Errorf("cursor store get: %w", err)
	}
	if !ok {
		cur = r.InitCursor
	}

	req := r.BuildRequest(cur, limit)

	resp, err := r.Scanner.ScanEvents(ctx, req)
	if err != nil {
		return fmt.Errorf("scan events: %w", err)
	}

	for _, ev := range resp.Events() {
		if err := r.Handle(ctx, ev); err != nil {
			return fmt.Errorf("handle event: %w", err)
		}
	}

	if err := r.Store.Set(ctx, resp.NextCursor()); err != nil {
		return fmt.Errorf("cursor store set: %w", err)
	}
	return nil
}
