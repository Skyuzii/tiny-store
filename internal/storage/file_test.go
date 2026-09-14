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
	writeTestFileStorage(t, path, `{"language":"go","editor":"vim"}`)

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
	writeTestFileStorage(t, path, "")

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
	writeTestFileStorage(t, path, contents)

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

func TestFileStoragePutExistingKeyOverwritesPersistedValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v, want nil", err)
	}
	if err := store.Put("language", "go"); err != nil {
		t.Fatalf("first Put() error = %v, want nil", err)
	}

	err = store.Put("language", "rust")
	if err != nil {
		t.Fatalf("second Put() error = %v, want nil", err)
	}

	reopened, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("reopen storage: %v", err)
	}
	got, err := reopened.Get("language")
	if err != nil {
		t.Fatalf("Get() after reopen error = %v, want nil", err)
	}
	if got != "rust" {
		t.Errorf("Get() after overwrite = %q, want %q", got, "rust")
	}
}

func TestFileStorageDeletePersistsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	writeTestFileStorage(t, path, `{"language":"go","editor":"vim"}`)

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

	entries := store.List()
	entries["language"] = "rust"
	entries["editor"] = "vim"

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

func writeTestFileStorage(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func TestFileStorageFailedDeleteDoesNotChangeMemory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	store, err := NewFileStorage(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put("language", "go"); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete("language"); err == nil {
		t.Fatal("expected persistence error")
	}
	if value, err := store.Get("language"); err != nil || value != "go" {
		t.Fatalf("Get() after failed Delete = %q, %v", value, err)
	}
}

func TestNewFileStorageAcceptsBlankAndNull(t *testing.T) {
	for _, contents := range []string{" \n\t", "null"} {
		t.Run(contents, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.json")
			writeTestFileStorage(t, path, contents)
			store, err := NewFileStorage(path)
			if err != nil {
				t.Fatal(err)
			}
			if entries := store.List(); len(entries) != 0 {
				t.Fatalf("List() = %v, want empty", entries)
			}
			if err := store.Put("", ""); err != nil {
				t.Fatal(err)
			}
			reopened, err := NewFileStorage(path)
			if err != nil {
				t.Fatal(err)
			}
			if value, err := reopened.Get(""); err != nil || value != "" {
				t.Fatalf("Get(empty) after reopen = %q, %v", value, err)
			}
		})
	}
}

func TestFileStorageFailedOverwritePreservesValue(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStorage(filepath.Join(dir, "data.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put("language", "go"); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := store.Put("language", "rust"); err == nil {
		t.Fatal("Put() error = nil, want persistence error")
	}
	if value, err := store.Get("language"); err != nil || value != "go" {
		t.Fatalf("Get() after failed overwrite = %q, %v, want go, nil", value, err)
	}
}
