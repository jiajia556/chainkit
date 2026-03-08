package types

import "time"

// Event is the unified output for "deposit listening" and reconciliation input.
type Event struct {
	Chain Chain
	Asset Asset

	TxHash    string
	Index     string // chain-specific unique index: evm logIndex, btc vout, tron event index...
	Height    uint64 // block height / slot
	Timestamp time.Time

	From string
	To   string

	Amount Amount

	// Raw is optional: store raw json / raw log bytes for debugging & reparse.
	Raw []byte

	// Tags: optional key/value for chain-specific metadata (topic0, method, memo, etc.)
	Tags map[string]string
}
