// Command generate writes each module's manifest and the listing of every shipped path, which
// package shipped builds from the shipped data. With -repin it writes the pins of the shipped
// shape and of the listing's shape instead.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/larvit/fejkdata/data"
	"github.com/larvit/fejkdata/internal/shipped"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	root := flags.String("root", ".", "the repository's root folder")
	repin := flags.Bool("repin", false, "write the pins under testdata/, and nothing else")
	if err := flags.Parse(args); errors.Is(err, flag.ErrHelp) {
		return nil
	} else if err != nil {
		return err
	}
	files, err := generated(*root, *repin)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(*root, filepath.FromSlash(f.Name)), f.Body, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func generated(root string, repin bool) ([]shipped.File, error) {
	loaded, err := shipped.Load(data.Modules())
	if err != nil {
		return nil, err
	}
	if repin {
		return []shipped.File{loaded.Shape(), shipped.ListingShape()}, nil
	}
	manifests, err := loaded.Manifests()
	if err != nil {
		return nil, err
	}
	kinds, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(shipped.KindsFile)))
	if err != nil {
		return nil, err
	}
	listed, err := loaded.Listing(kinds, shipped.ListingState)
	if err != nil {
		return nil, fmt.Errorf("the listing is not whole:\n%w", err)
	}
	return slices.Concat(manifests, loaded.ListingFiles(listed)), nil
}
