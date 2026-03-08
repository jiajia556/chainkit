package memory

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/jiajia556/chainkit/core/types"
	"github.com/jiajia556/chainkit/scan/cursor"
)

type Store struct {
	mu sync.RWMutex
	m  map[string]cursor.Cursor
}

func NewStore() *Store {
	return &Store{
		m: make(map[string]cursor.Cursor),
	}
}

func (s *Store) Get(ctx context.Context, chain types.Chain, scannerName string) (cursor.Cursor, bool, error) {
	_ = ctx
	if s == nil {
		return nil, false, fmt.Errorf("store is nil")
	}
	if scannerName == "" {
		return nil, false, fmt.Errorf("scannerName is empty")
	}

	key := storeKey(chain, scannerName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	cur, ok := s.m[key]
	if !ok {
		return nil, false, nil
	}

	out := make(cursor.Cursor, len(cur))
	for k, v := range cur {
		out[k] = v
	}
	return out, true, nil
}

func (s *Store) Set(ctx context.Context, chain types.Chain, scannerName string, cur cursor.Cursor) error {
	_ = ctx
	if s == nil {
		return fmt.Errorf("store is nil")
	}
	if scannerName == "" {
		return fmt.Errorf("scannerName is empty")
	}

	key := storeKey(chain, scannerName)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.m == nil {
		s.m = make(map[string]cursor.Cursor)
	}

	cp := make(cursor.Cursor, len(cur))
	for k, v := range cur {
		cp[k] = v
	}
	s.m[key] = cp
	return nil
}

func storeKey(c types.Chain, scannerName string) string {
	// type:network:chainID:name|scannerName
	var b strings.Builder
	b.WriteString(string(c.Type))
	b.WriteByte(':')
	b.WriteString(string(c.Network))
	b.WriteByte(':')
	b.WriteString(strconv.FormatUint(c.ChainID, 10))
	b.WriteByte(':')
	b.WriteString(c.Name)
	b.WriteByte('|')
	b.WriteString(scannerName)
	return b.String()
}
