package example

type Store interface {
	Update(key int64, value int64) error
	Get(key int64) (*int64, error)
	Delete(key int64) error
}
