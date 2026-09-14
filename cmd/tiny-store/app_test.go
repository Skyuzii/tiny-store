package main

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"tiny-store/internal/storage"
)

func TestAppInvalidCommandsDoNotStopSession(t *testing.T) {
	for _, command := range []string{"unknown", "put", "put key", "put key value extra", "get", "get key extra", "delete", "del", "delete key extra", "list extra", "exit extra"} {
		t.Run(command, func(t *testing.T) {
			var output, errorOutput bytes.Buffer
			app := &app{store: storage.NewMemoryStorage(), output: &output, errorOutput: &errorOutput}
			if err := app.run(strings.NewReader(command + "\nput valid value\nget valid\n")); err != nil {
				t.Fatal(err)
			}
			if errorOutput.Len() == 0 {
				t.Fatal("expected command error")
			}
			if !strings.HasSuffix(errorOutput.String(), "\n") {
				t.Fatal("error lacks newline")
			}
			if !strings.Contains(output.String(), "Got value value\n") {
				t.Fatalf("session did not recover: %q", output.String())
			}
		})
	}
}

func TestAppSession(t *testing.T) {
	var output, errorOutput bytes.Buffer
	app := &app{store: storage.NewMemoryStorage(), output: &output, errorOutput: &errorOutput}
	input := "\nput z last\nput a first\nput a changed\nget a\nlist\ndelete a\nget a\ndelete z\nexit\nput ignored value\n"
	if err := app.run(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Got value changed\n", "Key a with value changed\nKey z with value last\n", "Deleted key a\n", "Deleted key z\n"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("output = %q, want substring %q", output.String(), want)
		}
	}
	if got, want := errorOutput.String(), "key not found\n"; got != want {
		t.Errorf("errors = %q, want %q", got, want)
	}
	if entries := app.store.List(); len(entries) != 0 {
		t.Errorf("data after exit = %v, want empty", entries)
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestAppInputError(t *testing.T) {
	want := errors.New("input unavailable")
	app := &app{store: storage.NewMemoryStorage(), output: io.Discard, errorOutput: io.Discard}
	if err := app.run(failingReader{want}); !errors.Is(err, want) {
		t.Fatalf("run() = %v, want %v", err, want)
	}
}

func TestAppLongInput(t *testing.T) {
	app := &app{store: storage.NewMemoryStorage(), output: io.Discard, errorOutput: io.Discard}
	if err := app.run(strings.NewReader(strings.Repeat("x", 128*1024))); err == nil {
		t.Fatal("expected scanner limit error")
	}
}

func TestRunPersistsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	if err := run([]string{"-file", path}, strings.NewReader("put language go\n"), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run([]string{"-file", path}, strings.NewReader("get language\n"), &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Got value go\n") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRunStartup(t *testing.T) {
	for _, test := range []struct {
		name      string
		args      []string
		wantError bool
	}{
		{"help", []string{"-h"}, false},
		{"unknown flag", []string{"-unknown"}, true},
		{"positional argument", []string{"extra"}, true},
		{"unreadable storage", []string{"-file", t.TempDir()}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := run(test.args, strings.NewReader(""), io.Discard, io.Discard)
			if (err != nil) != test.wantError {
				t.Fatalf("run() = %v, wantError %v", err, test.wantError)
			}
		})
	}
}
