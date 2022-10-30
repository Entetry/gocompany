// Package cache contains local cache for different structs
package cache

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"entetry/gotest/internal/model"
)

// LocalCache cache company struct
type LocalCache struct {
	stop      chan struct{}
	mu        sync.RWMutex
	companies map[uuid.UUID]model.Company
}

var (
	errUserNotInCache = errors.New("the company isn't in cache")
)

// NewLocalCache creates new company cache object
func NewLocalCache() *LocalCache {
	lc := &LocalCache{
		companies: make(map[uuid.UUID]model.Company),
		stop:      make(chan struct{}),
	}

	return lc
}

// Update add or update entry to cache
func (lc *LocalCache) Update(ID uuid.UUID, name string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.companies[ID] = model.Company{ID: ID, Name: name}
}

// Read read entry from cache
func (lc *LocalCache) Read(id uuid.UUID) (*model.Company, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	cu, ok := lc.companies[id]
	if !ok {
		return nil, errUserNotInCache
	}

	return &cu, nil
}

// Delete remove entry from cache
func (lc *LocalCache) Delete(id uuid.UUID) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	delete(lc.companies, id)
}
