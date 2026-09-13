package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"tiny-store/internal/storage"
)

var ErrExit = errors.New("bye")

type CommandFunc func(args []string) error

type App struct {
	storage storage.Storage
}

func main() {
	fmt.Println("Tiny Store")
	fmt.Println("Commands: put <key> <value>, get <key>, delete <key>, list, exit")

	store, err := storage.NewFileStorage("data.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	app := &App{
		storage: store,
	}

	app.Run()
}

func (a *App) Run() {
	commands := map[string]CommandFunc{
		"put":  a.Put,
		"get":  a.Get,
		"del":  a.Delete,
		"list": a.List,
		"exit": a.Exit,
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		args := strings.Fields(line)

		if len(args) < 1 {
			continue
		}

		command := args[0]

		c, ok := commands[command]
		if !ok {
			fmt.Printf("Unknown command %s\n", command)
		}

		err := c(args[1:])

		if err != nil {
			fmt.Print(err)

			if errors.Is(err, ErrExit) {
				return
			}

			continue
		}
	}
}

func (a *App) Put(args []string) error {
	key := args[0]
	value := args[1]
	err := a.storage.Put(key, value)

	if err == nil {
		fmt.Printf("Put key %s with value %s\n", key, value)
	}

	return err
}

func (a *App) Get(args []string) error {
	key := args[0]
	v, err := a.storage.Get(key)

	if err == nil {
		fmt.Printf("Got value %s\n", v)
	}

	return err
}

func (a *App) Delete(args []string) error {
	key := args[0]
	err := a.storage.Delete(key)

	if err == nil {
		fmt.Printf("Deleted key %s\n", key)
	}

	return err
}

func (a *App) List([]string) error {
	res := a.storage.List()

	for key, value := range res {
		fmt.Printf("Key %s with value %s\n", key, value)
	}

	return nil
}

func (a *App) Exit([]string) error {
	return ErrExit
}
