package storage

import "errors"

var ErrNotFound = errors.New("key not found")

type Storage interface {
	Put(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	List() map[string]string
}
