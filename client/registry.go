package client

import (
	"sync"

	"github.com/jiajia556/chainkit/core/types"
)

type Adapter interface {
	Chain() types.Chain
}

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}
