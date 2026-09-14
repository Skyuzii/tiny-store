package storage

import (
	"errors"
	"maps"
	"testing"
)

func TestMemoryStoragePutAndGet(t *testing.T) {
	store := NewMemoryStorage()

	if err := store.Put("language", "go"); err != nil {
		t.Fatalf("Put() error = %v, want nil", err)
	}

	got, err := store.Get("language")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != "go" {
		t.Errorf("Get() = %q, want %q", got, "go")
	}
}

func TestMemoryStorageGetMissingKey(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.Get("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStorageDelete(t *testing.T) {
	store := NewMemoryStorage()
	if err := store.Put("language", "go"); err != nil {
		t.Fatalf("Put() error = %v, want nil", err)
	}

	if err := store.Delete("language"); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}

	_, err := store.Get("language")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after Delete() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStorageDeleteMissingKey(t *testing.T) {
	store := NewMemoryStorage()

	err := store.Delete("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStorageList(t *testing.T) {
	store := NewMemoryStorage()
	want := map[string]string{
		"language": "go",
		"editor":   "vim",
	}

	for key, value := range want {
		if err := store.Put(key, value); err != nil {
			t.Fatalf("Put(%q, %q) error = %v, want nil", key, value, err)
		}
	}

	if got := store.List(); !maps.Equal(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
}

func TestMemoryStorageListEmpty(t *testing.T) {
	store := NewMemoryStorage()

	if got := store.List(); len(got) != 0 {
		t.Errorf("List() = %v, want empty map", got)
	}
}

func TestMemoryStorageListReturnsCopy(t *testing.T) {
	store := NewMemoryStorage()
	if err := store.Put("language", "go"); err != nil {
		t.Fatalf("Put() error = %v, want nil", err)
	}

	result := store.List()
	result["language"] = "rust"
	result["editor"] = "vim"

	want := map[string]string{"language": "go"}
	if got := store.List(); !maps.Equal(got, want) {
		t.Errorf("List() after modifying returned map = %v, want %v", got, want)
	}
}
