package storage

import (
	"errors"
	"testing"
)

func TestMemoryStorage_PutGet(t *testing.T) {
	store := NewMemoryStorage()

	err := store.Put("key", "value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	v, err := store.Get("key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if v != "value" {
		t.Errorf("unexpected value: %v", v)
	}
}

func TestMemoryStorage_PutExistingKey(t *testing.T) {
	store := NewMemoryStorage()

	err := store.Put("key", "value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = store.Put("key", "value2")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMemoryStorage_GetMissingKey(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.Get("key")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	store := NewMemoryStorage()

	err := store.Put("key", "value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = store.Get("key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = store.Delete("key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = store.Get("key")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewMemoryStorage_List(t *testing.T) {
	store := NewMemoryStorage()

	expectedMap := map[string]string{
		"language": "go",
		"editor":   "vim",
	}

	for key, value := range expectedMap {
		if err := store.Put(key, value); err != nil {
			t.Fatalf("Put(%q): неожиданная ошибка: %v", key, err)
		}
	}

	got := store.List()

	if len(got) != len(expectedMap) {
		t.Fatalf("List: записей = %d, ожидалось %d", len(got), len(expectedMap))
	}

	for key, expectedValue := range expectedMap {
		gotValue, exists := got[key]
		if !exists {
			t.Errorf("List: отсутствует ключ %q", key)
			continue
		}

		if gotValue != expectedValue {
			t.Errorf("List[%q] = %q, ожидалось %q", key, gotValue, expectedValue)
		}
	}
}
