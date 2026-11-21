package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/larvit/fejkdata"
)

func fail(stderr io.Writer, name string, err error) int {
	var te inputError
	if errors.As(err, &te) {
		return misuse(stderr, name, te.error)
	}
	report(stderr, name, err)
	if errors.Is(err, fejkdata.ErrNoColumns) {
		fmt.Fprintln(stderr, "wrap that JSON in single quotes, which keep a shell from expanding its braces")
	}
	return 1
}

func misuse(stderr io.Writer, name string, err error) int {
	report(stderr, name, err)
	fmt.Fprintf(stderr, "try '%s --help'\n", name)
	return 2
}

// report writes err under the command name, in place of the library's own on each line.
func report(stderr io.Writer, name string, err error) {
	msg := strings.ReplaceAll(strings.TrimPrefix(err.Error(), "fejkdata: "), "\nfejkdata: ", "\n"+name+": ")
	fmt.Fprintf(stderr, "%s: %s\n", name, msg)
}
