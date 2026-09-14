package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"tiny-store/internal/storage"
)

var errExit = errors.New("exit")

type commandFunc func(args []string) error

type app struct {
	store       storage.Storage
	output      io.Writer
	errorOutput io.Writer
}

func (a *app) run(input io.Reader) error {
	commands := map[string]commandFunc{
		"put":    a.put,
		"get":    a.get,
		"delete": a.delete,
		"list":   a.list,
		"exit":   a.exit,
	}

	_, _ = fmt.Fprintln(a.output, "Tiny Store")
	_, _ = fmt.Fprintln(a.output, "Commands: put <key> <value>, get <key>, delete <key>, list, exit")
	scanner := bufio.NewScanner(input)
	for {
		_, _ = fmt.Fprint(a.output, "> ")
		if !scanner.Scan() {
			break
		}
		args := strings.Fields(scanner.Text())
		if len(args) == 0 {
			continue
		}
		command := args[0]
		handler, ok := commands[command]
		if !ok {
			_, _ = fmt.Fprintf(a.errorOutput, "unknown command %q\n", command)
			continue
		}
		err := handler(args[1:])
		if errors.Is(err, errExit) {
			return nil
		}
		if err != nil {
			_, _ = fmt.Fprintln(a.errorOutput, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	return nil
}

func (a *app) put(args []string) error {
	if len(args) != 2 {
		return errors.New("usage: put <key> <value>")
	}
	key, value := args[0], args[1]
	if err := a.store.Put(key, value); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(a.output, "Put key %s with value %s\n", key, value)
	return nil
}

func (a *app) get(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: get <key>")
	}
	key := args[0]
	value, err := a.store.Get(key)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(a.output, "Got value %s\n", value)
	return nil
}

func (a *app) delete(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: delete <key>")
	}
	key := args[0]
	if err := a.store.Delete(key); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(a.output, "Deleted key %s\n", key)
	return nil
}

func (a *app) list(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: list")
	}
	entries := a.store.List()
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		_, _ = fmt.Fprintf(a.output, "Key %s with value %s\n", key, entries[key])
	}
	return nil
}

func (a *app) exit(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: exit")
	}
	return errExit
}
