package registry

// Registry provides an interface for managing clients

type Registry interface {
    Register(client Client) error
    Get(id string) (Client, error)
}

// NewRegistry creates a new client registry
func NewRegistry() Registry {
    return &registryImpl{clients: make(map[string]Client)}
}

type registryImpl struct {
    clients map[string]Client
}

// Register adds a new client to the registry
func (r *registryImpl) Register(client Client) error {
    if _, exists := r.clients[client.ID]; exists {
        return fmt.Errorf("client already registered: %s", client.ID)
    }
    r.clients[client.ID] = client
    return nil
}

// Get retrieves a client by id
func (r *registryImpl) Get(id string) (Client, error) {
    client, exists := r.clients[id]
    if !exists {
        return Client{}, fmt.Errorf("client not found: %s", id)
    }
    return client, nil
}