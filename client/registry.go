// registry.go

package client

import (
    "sync"
)

// Adapter is the existing interface
type Adapter interface {
    // existing methods
}

// Registry struct definition with the internal adapters map and mutex
type Registry struct {
    mu       sync.Mutex
    adapters map[string]Adapter
}

// NewRegistry creates a new Registry instance
func NewRegistry() *Registry {
    return &Registry{
        adapters: make(map[string]Adapter),
    }
}

// Register adds a new adapter to the Registry
func (r *Registry) Register(adapter Adapter) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // registration logic
}

// Get retrieves an adapter by Chain type
func (r *Registry) Get(chain types.Chain) (Adapter, bool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    // retrieval logic
}
