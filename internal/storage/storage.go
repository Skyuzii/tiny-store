package storage

type Storage interface {
	Put(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	List() map[string]string
}
