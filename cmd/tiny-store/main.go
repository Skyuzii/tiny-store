package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"tiny-store/internal/storage"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, input io.Reader, output, errorOutput io.Writer) error {
	flags := flag.NewFlagSet("tiny-store", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	path := flags.String("file", "data.json", "path to the storage JSON file (parent directory must exist)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	store, err := storage.NewFileStorage(*path)
	if err != nil {
		return err
	}
	app := &app{store: store, output: output, errorOutput: errorOutput}
	return app.run(input)
}
