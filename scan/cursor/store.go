package cursor

import "context"

// Store is a generic interface for persisting and retrieving a cursor.
// Cur is the chain-specific cursor type.
// Get/Set have no chain or scannerName arguments; each Store instance
// is scoped to a single scanner.
type Store[Cur any] interface {
	Get(ctx context.Context) (Cur, bool, error)
	Set(ctx context.Context, cur Cur) error
}
