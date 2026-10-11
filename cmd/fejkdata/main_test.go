package main

import (
	"io"
	"os"
	"testing"

	"github.com/larvit/fejkdata/cli"
	"github.com/larvit/fejkdata/data"
)

func TestMainRendersTheShippedData(t *testing.T) {
	stdin, feed, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	drain, stdout, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	args, oldIn, oldOut := os.Args, os.Stdin, os.Stdout
	t.Cleanup(func() { os.Args, os.Stdin, os.Stdout = args, oldIn, oldOut })
	os.Args, os.Stdin, os.Stdout = []string{"fejkdata", "--seed", "1"}, stdin, stdout
	io.WriteString(feed, "{/misc.territory[SE].alpha2} {/en_US.person.last} {/sv_SE.address.locality}")
	feed.Close()
	code := cli.Main(data.Modules()...)
	stdout.Close()
	out, _ := io.ReadAll(drain)
	if code != 0 || len(out) < len("SE x y") || string(out[:3]) != "SE " {
		t.Errorf("fejkdata = %d, %q, want SE followed by a last name and a locality", code, out)
	}
}
