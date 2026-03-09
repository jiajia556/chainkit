package tron

// Cursor tracks the scanning position on a TRON chain.
type Cursor struct {
	BlockNumber uint64
}

// Target describes the contract and event to scan on TRON.
type Target struct {
	ContractAddress string
	EventName       string
}

// ScanRequest is the input to a TRON event scanner.
type ScanRequest struct {
	Target        Target
	Cursor        Cursor
	Confirmations uint64
	Limit         int
}

// Event is a single scanned TRON contract event.
type Event struct {
	TxHash      string
	BlockNumber uint64
	EventName   string
	Result      map[string]string
}

// ScanResponse is the output of a TRON event scanner.
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
