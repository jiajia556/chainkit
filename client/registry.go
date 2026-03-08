package client

import (
	"sync"
	"github.com/jiajia556/chainkit/core/types"
)

// Adapter interface defines the methods that must be implemented by any adapter.
type Adapter interface {
	Chain() types.Chain
}

// Registry holds a map of adapters and a mutex for concurrent access.
type Registry struct {
	mu      sync.RWMutex
	adapters map[chainKey]Adapter
}
// NewRegistry creates a new instance of Registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[chainKey]Adapter),
	}
}

