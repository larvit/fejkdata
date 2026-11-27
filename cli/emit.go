package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata"
)

func parseEmit(v string) (fejkdata.Emit, error) {
	template, count, counted := strings.Cut(v, "=")
	if template == "" {
		return fejkdata.Emit{}, fmt.Errorf("--emit needs <template>=<count>, got %q", v)
	}
	if !counted {
		return fejkdata.Emit{}, fmt.Errorf("--emit %s gives no count; write --emit %s=<count>", v, v)
	}
	n, err := strconv.Atoi(count)
	if err != nil || n < 0 || n > fejkdata.MaxRepeat {
		return fejkdata.Emit{}, fmt.Errorf("--emit %s needs a count in 0..%d, got %q", template, fejkdata.MaxRepeat, count)
	}
	return fejkdata.Emit{Template: template, Count: n}, nil
}

// checkEmitFlags rejects --out-dir without --emit, and --emit beside a flag that sets what each
// template sets for itself, or under a format that can't hold its templates.
func (in invocation) checkEmitFlags() error {
	switch {
	case in.list:
		return nil
	case len(in.emits) == 0 && in.outDirSet:
		return errors.New("--out-dir names the folder --emit writes to, and no --emit is given")
	case len(in.emits) == 0:
		return nil
	case in.tableSet:
		return errors.New("--table names one target, and under --emit each template names its own; drop --table")
	case in.repeatSet:
		return fmt.Errorf("--repeat counts one template's records, and under --emit each template has its own count, as in --emit %s=%d; drop --repeat", in.emits[0].Template, in.emits[0].Count)
	}
	return in.checkEmitFormat()
}

func (in invocation) checkEmitFormat() error {
	f, records := recordFormats[in.format]
	switch {
	case !records:
		return errors.New("--emit writes records, so it needs --format json, ndjson, csv, tsv, pgcopy or sql")
	case in.outDirSet && in.outDir == "":
		return errors.New("--out-dir names a folder, so it cannot be empty")
	case len(in.emits) > 1 && f.target == "" && !in.outDirSet:
		return fmt.Errorf("--format %s cannot hold more than one template in one stream; give --out-dir <folder>, which gets a file per template", in.format)
	}
	return nil
}

// emit writes each --emit template in turn. A run the library refuses with ErrRun is misuse.
func (in invocation) emit(modules []fs.FS, sys system, stdin io.Reader, stdout io.Writer) error {
	if stdin != nil {
		raw, err := io.ReadAll(stdin)
		if err != nil {
			return inputError{fmt.Errorf("stdin cannot be read: %w", err)}
		}
		if len(raw) > 0 {
			return inputError{errors.New("--emit names its templates, and stdin holds one too; send nothing on stdin, or drop --emit")}
		}
	}
	format := in.recordFormat()
	err := fejkdata.FakeRun(in.emits, func(e fejkdata.Emit, fake func() (*fejkdata.Record, error)) error {
		return in.writeTemplate(e, fake, format, stdout)
	}, in.options(modules, sys)...)
	if errors.Is(err, fejkdata.ErrRun) {
		return inputError{fmt.Errorf("--emit: %s", strings.TrimPrefix(err.Error(), "fejkdata: "))}
	}
	return err
}

// writeTemplate writes e's records to stdout, or under --out-dir to the file named by e's template and the format.
func (in invocation) writeTemplate(e fejkdata.Emit, fake recordSource, format recordFormat, stdout io.Writer) (err error) {
	w := stdout
	if in.outDirSet {
		if err := os.MkdirAll(in.outDir, 0o777); err != nil {
			return err
		}
		file, err := os.Create(filepath.Join(in.outDir, e.Template+"."+in.format))
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, file.Close()) }()
		w = file
	}
	if e.Count == 0 {
		return writeNoRecord(fake, format, w)
	}
	return format.write(fake, e.Count, fejkdata.PathName(e.Template), bufio.NewWriter(w))
}
