package ifchanged

type DB interface {
	Put(key, value []byte) error
	Has(key []byte) bool
	Get(key []byte) ([]byte, error)
	Sync() error
	Close() error
}
