package types

type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
	TxStatusNotFound  TxStatus = "not_found" // dropped or unknown
)

type TxResult struct {
	Chain  Chain
	Hash   string
	Status TxStatus
	Height uint64
	// Confirmations optional: HeightLatest - Height + 1
}
