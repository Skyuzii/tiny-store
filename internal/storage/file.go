package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"
)

type FileStorage struct {
	mu   sync.RWMutex
	path string
	data map[string]string
}

func NewFileStorage(path string) (*FileStorage, error) {
	data := make(map[string]string)

	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &FileStorage{
			path: path,
			data: data,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read storage file %q: %w", path, err)
	}

	if len(bytes.TrimSpace(contents)) > 0 {
		if err := json.Unmarshal(contents, &data); err != nil {
			return nil, fmt.Errorf("decode storage file %q: %w", path, err)
		}
		if data == nil {
			data = make(map[string]string)
		}
	}

	return &FileStorage{
		path: path,
		data: data,
	}, nil
}

func (s *FileStorage) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nextData := maps.Clone(s.data)
	nextData[key] = value

	if err := s.save(nextData); err != nil {
		return err
	}

	s.data = nextData

	return nil
}

func (s *FileStorage) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[key]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

func (s *FileStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return ErrNotFound
	}

	nextData := maps.Clone(s.data)
	delete(nextData, key)

	if err := s.save(nextData); err != nil {
		return err
	}

	s.data = nextData

	return nil
}

func (s *FileStorage) List() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return maps.Clone(s.data)
}

func (s *FileStorage) save(data map[string]string) error {
	contents, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode storage data: %w", err)
	}
	contents = append(contents, '\n')

	dir := filepath.Dir(s.path)
	tempFile, err := os.CreateTemp(dir, "."+filepath.Base(s.path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary storage file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if err := tempFile.Chmod(0o644); err != nil {
		return fmt.Errorf("set storage file permissions: %w", err)
	}
	if _, err := tempFile.Write(contents); err != nil {
		return fmt.Errorf("write storage file: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("sync storage file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close storage file: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace storage file: %w", err)
	}

	return nil
}
