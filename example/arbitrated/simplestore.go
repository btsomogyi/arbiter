package arbitrated

import (
	"github.com/btsomogyi/arbiter/example"
)

var _ example.Store = (*SimpleStore)(nil)

// SimpleStore is a stand-in for any data backend; Element key / version map.
type SimpleStore map[int64]int64

// NewSimpleStore creates an initialized SimpleStore.
func NewSimpleStore() SimpleStore {
	return make(map[int64]int64)
}

// Update sets the installed version for the passed key in SimpleStore to the
// version provided.  In a less trivial implementation, this would be a persistent
// data access function (DBMS, keystore, etc).
func (db SimpleStore) Update(key int64, version int64) error {
	db[key] = version

	return nil
}

// Get returns the current version of a given Element key, returning nil if not found.
func (db SimpleStore) Get(key int64) (*int64, error) {
	v, ok := db[key]
	if !ok {
		return nil, example.ErrKeyNotFound
	}

	return &v, nil
}

func (db SimpleStore) Delete(key int64) error {
	_, ok := db[key]
	if !ok {
		return example.ErrKeyNotFound
	}
	delete(db, key)
	return nil
}
