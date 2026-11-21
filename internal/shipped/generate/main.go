// Command generate writes the files package shipped generates from the shipped data: each
// module's manifest and the listing of every shipped path, or with -repin the pins of the
// shipped shape and of the listing's shape instead, so a re-pin is its own act.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/larvit/fejkdata/data"
	"github.com/larvit/fejkdata/internal/shipped"
)

func main() {
	root := flag.String("root", ".", "the repository's root folder")
	repin := flag.Bool("repin", false, "write the pins under testdata/, and nothing else")
	flag.Parse()
	files, err := generated(*root, *repin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(*root, filepath.FromSlash(f.Name)), f.Body, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "generate:", err)
			os.Exit(1)
		}
	}
}

func generated(root string, repin bool) ([]shipped.File, error) {
	tree, namespaces, err := shipped.Load(data.Modules())
	if err != nil {
		return nil, err
	}
	if repin {
		return []shipped.File{shipped.Shape(&tree), shipped.ListingShape()}, nil
	}
	manifests, err := shipped.Manifests(&tree, namespaces)
	if err != nil {
		return nil, err
	}
	kinds, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(shipped.KindsFile)))
	if err != nil {
		return nil, err
	}
	listed, err := shipped.Listing(&tree, kinds, shipped.ListingState)
	if err != nil {
		return nil, fmt.Errorf("the listing is not whole:\n%w", err)
	}
	return slices.Concat(manifests, shipped.ListingFiles(listed, namespaces)), nil
}
