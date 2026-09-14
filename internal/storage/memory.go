package storage

import (
	"maps"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (s *MemoryStorage) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	return nil
}

func (s *MemoryStorage) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[key]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

func (s *MemoryStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return ErrNotFound
	}

	delete(s.data, key)
	return nil
}

func (s *MemoryStorage) List() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return maps.Clone(s.data)
}
