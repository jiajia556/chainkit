package ports

import "context"

// ScanPort is the generic interface for chain event scanners.
// Req is the chain-specific scan request type.
// Resp is the chain-specific scan response type.
type ScanPort[Req any, Resp any] interface {
	ScanEvents(ctx context.Context, req Req) (Resp, error)
}
