// Command fejkdata renders the template on its stdin with fake values from every shipped
// module; package cli documents it.
package main

import (
	"os"

	"github.com/larvit/fejkdata/cli"
	"github.com/larvit/fejkdata/data"
)

func main() { os.Exit(cli.Main(data.Modules()...)) }
