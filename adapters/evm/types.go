package evm

// Cursor tracks the scanning position on an EVM chain.
type Cursor struct {
	BlockNumber uint64
	LogIndex    uint
}

// Target describes the contract and event topic to scan.
type Target struct {
	ContractAddress string
	Topic0          string // keccak256 of the event signature
}

// ScanRequest is the input to an EVM event scanner.
type ScanRequest struct {
	Target        Target
	Cursor        Cursor
	Confirmations uint64
	Limit         int
}

// Event is a single scanned EVM log entry.
type Event struct {
	TxHash      string
	BlockNumber uint64
	LogIndex    uint
	Address     string
	Topics      []string
	Data        []byte
}

// ScanResponse is the output of an EVM event scanner.
// It satisfies runner.Response[Event, Cursor] via Events() and NextCursor().
type ScanResponse struct {
	events     []Event
	nextCursor Cursor
}

// NewScanResponse constructs a ScanResponse.
func NewScanResponse(events []Event, nextCursor Cursor) ScanResponse {
	return ScanResponse{events: events, nextCursor: nextCursor}
}

// Events returns the scanned events.
func (r ScanResponse) Events() []Event { return r.events }

// NextCursor returns the cursor to persist after successful handling.
func (r ScanResponse) NextCursor() Cursor { return r.nextCursor }
