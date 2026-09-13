package storage

import (
	"errors"
)

var ErrNotFound = errors.New("key not found")
var ErrAlreadyExists = errors.New("key already exists")

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (s *MemoryStorage) Put(key, value string) error {
	if _, ok := s.data[key]; ok {
		return ErrAlreadyExists
	}

	s.data[key] = value
	return nil
}

func (s *MemoryStorage) Get(key string) (string, error) {
	v, ok := s.data[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (s *MemoryStorage) Delete(key string) error {
	_, ok := s.data[key]
	if !ok {
		return ErrNotFound
	}

	delete(s.data, key)
	return nil
}

func (s *MemoryStorage) List() map[string]string {
	result := make(map[string]string, len(s.data))

	for key, value := range s.data {
		result[key] = value
	}

	return result
}
