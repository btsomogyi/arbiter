package locking

import (
	"github.com/btsomogyi/arbiter/example"
	"sync"
)

var _ example.Store = (*LockingStore)(nil)

// LockingStore is a stand-in for any data backend; Element key / version map protected
// by a simple mutex lock.
type LockingStore struct {
	elements map[int64]int64
	mtx      sync.Mutex
}

// NewElementStore creates an initialized LockingStore.
func NewElementStore() LockingStore {
	return LockingStore{
		elements: make(map[int64]int64),
		mtx:      sync.Mutex{},
	}
}

// Update obtains the LockingStore lock then adds/updates the key to the value provided.
func (e *LockingStore) Update(key int64, value int64) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	e.elements[key] = value
	return nil
}

// Get returns the value at the key from LockingStore once the lock is obtained.
func (e *LockingStore) Get(key int64) (*int64, error) {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	v, ok := e.elements[key]
	if !ok {
		return nil, example.ErrKeyNotFound
	}
	return &v, nil
}

// Delete removes an entry from the LockingStore after obtaining the mutex lock.
func (e *LockingStore) Delete(key int64) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	delete(e.elements, key)
	return nil
}
