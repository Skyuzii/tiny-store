package storage

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileStorageMissingFileCreatesEmptyStorage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")

	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}
	if got := store.List(); len(got) != 0 {
		t.Errorf("List() = %v, want empty map", got)
	}
}

func TestNewFileStorageLoadsExistingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	writeTestFile(t, path, `{"language":"go","editor":"vim"}`)

	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}

	got := store.List()
	if got["language"] != "go" || got["editor"] != "vim" || len(got) != 2 {
		t.Errorf("List() = %v, want map[language:go editor:vim]", got)
	}
}

func TestNewFileStorageEmptyFileCreatesEmptyStorage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	writeTestFile(t, path, "")

	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}
	if got := store.List(); len(got) != 0 {
		t.Errorf("List() = %v, want empty map", got)
	}
}

func TestNewFileStorageRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	const contents = `{not valid json`
	writeTestFile(t, path, contents)

	if _, err := NewFileStorage(path); err == nil {
		t.Fatal("NewFileStorage() error = nil, want invalid JSON error")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v, want nil", err)
	}
	if string(got) != contents {
		t.Errorf("file contents after failed open = %q, want %q", got, contents)
	}
}

func TestFileStoragePutPersistsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}

	if err := store.Put("language", "go"); err != nil {
		t.Fatalf("Put() error = %v, want nil", err)
	}

	reopened, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}
	got, err := reopened.Get("language")
	if err != nil {
		t.Fatalf("Get() after reopen error = %v, want nil", err)
	}
	if got != "go" {
		t.Errorf("Get() after reopen = %q, want %q", got, "go")
	}
}

func TestFileStoragePutExistingKeyPreservesPersistedValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}
	if err := store.Put("language", "go"); err != nil {
		t.Fatalf("first Put() error = %v, want nil", err)
	}

	err = store.Put("language", "rust")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("second Put() error = %v, want ErrAlreadyExists", err)
	}

	reopened, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}
	got, err := reopened.Get("language")
	if err != nil {
		t.Fatalf("Get() after reopen error = %v, want nil", err)
	}
	if got != "go" {
		t.Errorf("Get() after rejected Put() = %q, want %q", got, "go")
	}
}

func TestFileStorageDeletePersistsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	writeTestFile(t, path, `{"language":"go","editor":"vim"}`)

	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}
	if err := store.Delete("language"); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}

	reopened, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}
	if _, err := reopened.Get("language"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(deleted key) error = %v, want ErrNotFound", err)
	}
	got, err := reopened.Get("editor")
	if err != nil {
		t.Fatalf("Get(remaining key) error = %v, want nil", err)
	}
	if got != "vim" {
		t.Errorf("Get(remaining key) = %q, want %q", got, "vim")
	}
}

func TestFileStorageDeleteMissingKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}

	err = store.Delete("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestFileStorageListReturnsCopy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}
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

func TestFileStorageFailedPutDoesNotChangeMemory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}
	if err := store.Put("language", "go"); err == nil {
		t.Fatal("Put() error = nil, want persistence error")
	}

	if _, err := store.Get("language"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after failed Put() error = %v, want ErrNotFound", err)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
