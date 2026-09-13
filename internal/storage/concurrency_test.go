package storage

import (
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

const concurrentWorkers = 32

func TestMemoryStorageConcurrentAccess(t *testing.T) {
	testConcurrentAccess(t, NewMemoryStorage())
}

func TestFileStorageConcurrentAccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}

	testConcurrentAccess(t, store)

	reopened, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}
	if got := reopened.List(); len(got) != 0 {
		t.Errorf("reopened List() = %v, want empty map", got)
	}
}

func testConcurrentAccess(t *testing.T, store Storage) {
	t.Helper()

	var wg sync.WaitGroup
	errorsCh := make(chan error, concurrentWorkers)

	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			key := strconv.Itoa(i)
			if err := store.Put(key, "value"); err != nil {
				errorsCh <- fmt.Errorf("Put(%q): %w", key, err)
				return
			}

			if _, err := store.Get(key); err != nil {
				errorsCh <- fmt.Errorf("Get(%q): %w", key, err)
			}

			store.List()
		}(i)
	}

	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		t.Error(err)
	}

	if got := len(store.List()); got != concurrentWorkers {
		t.Fatalf("List() length after concurrent Put() = %d, want %d", got, concurrentWorkers)
	}

	errorsCh = make(chan error, concurrentWorkers)
	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			key := strconv.Itoa(i)
			if err := store.Delete(key); err != nil {
				errorsCh <- fmt.Errorf("Delete(%q): %w", key, err)
			}

			store.List()
		}(i)
	}

	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		t.Error(err)
	}

	if got := store.List(); len(got) != 0 {
		t.Errorf("List() after concurrent Delete() = %v, want empty map", got)
	}
}
