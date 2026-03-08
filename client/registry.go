package client

import "github.com/jiajia556/chainkit/core/types"

type Adapter interface {
	Chain() types.Chain
}

type Registry struct {
	// map[chainKey]Adapter
}
