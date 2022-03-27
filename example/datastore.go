package example

import "sync"

// ElementStore is a stand-in for any data backend; Element key / version map protected
// by a simple mutex lock.
type ElementStore struct {
	elements map[int64]int64
	mtx      sync.Mutex
}

// NewElementStore creates an initialized ElementStore.
func NewElementStore() ElementStore {
	return ElementStore{
		elements: make(map[int64]int64),
		mtx:      sync.Mutex{},
	}
}

// Update obtains the ElementStore lock then adds/updates the key to the value provided.
func (e *ElementStore) Update(key int64, value int64) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	if e.elements == nil {
		e.elements = make(map[int64]int64)
	}
	e.elements[key] = value
	return nil
}

// Get returns the value at the key from ElementStore once the lock is obtained.
func (e *ElementStore) Get(key int64) (int64, bool) {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	if e.elements == nil {
		e.elements = make(map[int64]int64)
	}
	v, ok := e.elements[key]
	return v, ok
}

// Delete removes an entry from the ElementStore after obtaining the mutex lock.
func (e *ElementStore) Delete(key int64) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	if e.elements == nil {
		e.elements = make(map[int64]int64)
	}
	delete(e.elements, key)
	return nil
}
