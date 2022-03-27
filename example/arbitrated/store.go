package arbitrated

import "fmt"

// ElementStore is a stand-in for any data backend; Element key / version map.
type ElementStore map[int64]int64

// ErrDbKeyNotFound indicates a failure to find the requested key in DB.
var ErrDbKeyNotFound = fmt.Errorf("key not found")

// setVersion sets the installed version for the passed key in ElementStore to the
// version provided.  In a less trivial implementation, this would be a persistent
// data access function (DBMS, keystore, etc).
func (db ElementStore) setVersion(key int64, version int64) error {
	db[key] = version

	return nil
}

// getVersion returns the current version of a given Element key, returning nil if not found.
func (db ElementStore) getVersion(key int64) (int64, error) {
	v, ok := db[key]
	if !ok {
		return v, ErrDbKeyNotFound
	}

	return v, nil
}
