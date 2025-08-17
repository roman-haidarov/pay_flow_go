package cache

type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
	Close() error
}
