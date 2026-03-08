package client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jiajia556/chainkit/core/types"
)

// NewRegistry returns a new Registry instance.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]Adapter),
	}
}

// Register adds an adapter to the Registry.
// Duplicate registration (same chain key) returns an error.
func (r *Registry) Register(adapter Adapter) error {
	if r == nil {
		return fmt.Errorf("registry is nil")
	}
	if adapter == nil {
		return fmt.Errorf("adapter is nil")
	}

	ch := adapter.Chain()
	key := chainKey(ch)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.adapters == nil {
		r.adapters = make(map[string]Adapter)
	}
	if _, ok := r.adapters[key]; ok {
		return fmt.Errorf("adapter already registered for chain: %s", key)
	}

	r.adapters[key] = adapter
	return nil
}

// Get retrieves an adapter for a given chain.
func (r *Registry) Get(chain types.Chain) (Adapter, bool) {
	if r == nil {
		return nil, false
	}

	key := chainKey(chain)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.adapters == nil {
		return nil, false
	}
	a, ok := r.adapters[key]
	return a, ok
}

func chainKey(c types.Chain) string {
	// Stable key: type:network:chainID:name
	// name kept to avoid collisions when chainID is 0 for non-EVM chains.
	var b strings.Builder
	b.WriteString(string(c.Type))
	b.WriteByte(':')
	b.WriteString(string(c.Network))
	b.WriteByte(':')
	b.WriteString(strconv.FormatUint(c.ChainID, 10))
	b.WriteByte(':')
	b.WriteString(c.Name)
	return b.String()
}
