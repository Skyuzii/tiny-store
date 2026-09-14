package storage

import (
	"errors"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

func TestStorageContract(t *testing.T) {
	factories := map[string]func(*testing.T) Storage{
		"memory": func(t *testing.T) Storage { return NewMemoryStorage() },
		"file": func(t *testing.T) Storage {
			t.Helper()
			store, err := NewFileStorage(filepath.Join(t.TempDir(), "data.json"))
			if err != nil {
				t.Fatal(err)
			}
			return store
		},
	}
	for name, newStore := range factories {
		t.Run(name, func(t *testing.T) {
			t.Run("empty strings", func(t *testing.T) {
				store := newStore(t)
				if _, err := store.Get("missing"); !errors.Is(err, ErrNotFound) {
					t.Fatalf("Get() = %v, want ErrNotFound", err)
				}
				if err := store.Put("", ""); err != nil {
					t.Fatal(err)
				}
				if value, err := store.Get(""); err != nil || value != "" {
					t.Fatalf("Get(empty) = %q, %v", value, err)
				}
				if err := store.Put("", "replacement"); err != nil {
					t.Fatalf("Put() = %v, want nil", err)
				}
				if value, err := store.Get(""); err != nil || value != "replacement" {
					t.Fatalf("Get after overwrite = %q, %v", value, err)
				}
				if err := store.Delete(""); err != nil {
					t.Fatal(err)
				}
			})
			t.Run("concurrent same key", func(t *testing.T) {
				store := newStore(t)
				errorsCh := make(chan error, concurrentWorkers)
				start := make(chan struct{})
				var wg sync.WaitGroup
				for i := 0; i < concurrentWorkers; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						<-start
						value := strconv.Itoa(i)
						errorsCh <- store.Put("shared", value)
					}()
				}
				close(start)
				wg.Wait()
				close(errorsCh)
				for err := range errorsCh {
					if err != nil {
						t.Errorf("Put() = %v, want nil", err)
					}
				}
				value, err := store.Get("shared")
				if err != nil {
					t.Fatal(err)
				}
				index, err := strconv.Atoi(value)
				if err != nil || index < 0 || index >= concurrentWorkers {
					t.Fatalf("Get() = %q, want one of the written values", value)
				}
				if entries := store.List(); len(entries) != 1 {
					t.Fatalf("List() = %v, want one key", entries)
				}

			})
		})
	}
}
