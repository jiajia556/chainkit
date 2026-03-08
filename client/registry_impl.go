package client

// NewRegistry returns a new Registry instance.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds an adapter to the Registry.
func (r *Registry) Register(adapter Adapter) {
	// Add logic to register adapter
}

// Get retrieves an adapter for a given chain.
func (r *Registry) Get(chain types.Chain) (Adapter, bool) {
	// Add logic to get adapter for chain
	return nil, false
}