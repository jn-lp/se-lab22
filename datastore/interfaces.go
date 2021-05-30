package datastore

type Reader interface {
	Get(key string) ([]byte, error)
	Has(key string) (bool, error)
}

type Writer interface {
	Put(key string, value []byte) error
	Delete(key string) error
}
