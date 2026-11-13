package cli

import (
	"bufio"
	"io"
	"sort"
	"strings"

	"github.com/larvit/fejkdata"
)

type recordSource func() (*fejkdata.Record, error)

// recordFormat is one --format that writes records. write streams them, and empty is what a repeat of 0 prints.
type recordFormat struct {
	writes string // what it writes between values, which --separator would seem to set
	write  func(record recordSource, repeat int, table string, out *bufio.Writer) error
	empty  func(*fejkdata.Record) string // nil where zero records print nothing
	target string                        // what --table names under this format, "" where it does nothing
}

// SQL Server refuses a VALUES list above 1000 rows (error 10738).
const sqlStatementRows = 1000

// recordFormats is every --format that writes records.
var recordFormats = map[string]recordFormat{
	"csv": {
		writes: `always separates fields with ","`,
		write:  recordLines{header: (*fejkdata.Record).CSVHeader, line: (*fejkdata.Record).CSVLine, sep: "\n"}.write,
		empty:  func(r *fejkdata.Record) string { return r.CSVHeader() + "\n" },
	},
	"json": {
		writes: "writes the records as one JSON array",
		write:  recordLines{line: (*fejkdata.Record).JSON, open: "[", close: "]", sep: ",\n"}.write,
		empty:  func(*fejkdata.Record) string { return "[]\n" },
	},
	"ndjson": {
		writes: "writes one record per line",
		write:  recordLines{line: (*fejkdata.Record).JSON, sep: "\n"}.write,
	},
	"tsv": {
		writes: "always separates fields with a tab",
		write:  recordLines{header: (*fejkdata.Record).TSVHeader, line: (*fejkdata.Record).TSVLine, sep: "\n"}.write,
		empty:  func(r *fejkdata.Record) string { return r.TSVHeader() + "\n" },
	},
	"pgcopy": {writes: "always separates fields with a tab", write: writePGCopy, target: "COPY"},
	"sql":    {writes: "writes one INSERT per line", write: writeSQL, target: "INSERT"},
}

// formatNames lists the --format values, text included, for the misuse error.
func formatNames() string {
	names := []string{"text"}
	for name := range recordFormats {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]
}

// recordStream builds the record drawer for the input, plus the INSERT table
// a sql format names.
func (in invocation) recordStream(f *fejkdata.Generator, src input) (func() (*fejkdata.Record, error), string, error) {
	record := func() (*fejkdata.Record, error) { return f.FakeRecord(src.record) }
	if src.record == "" {
		t, err := f.NewRecordTemplate(src.template)
		if err != nil {
			return nil, "", templateError(spacedReference(src.template, err))
		}
		record = func() (*fejkdata.Record, error) { return t.Fake(), nil }
	}
	table := in.table
	if table == "" {
		table = defaultTable(src.record)
	}
	return record, table, nil
}

func (in invocation) writeRecords(f *fejkdata.Generator, src input, w io.Writer) error {
	record, table, err := in.recordStream(f, src)
	if err != nil {
		return err
	}
	format := recordFormats[in.format]
	if in.repeat == 0 {
		return writeNoRecord(record, format, w)
	}
	return format.write(record, in.repeat, table, bufio.NewWriter(w))
}

// recordLines writes each record on a line of its own, framed by open and close, with the header
// before the first where the format has one.
type recordLines struct {
	header func(*fejkdata.Record) string
	line   func(*fejkdata.Record) string
	open   string
	close  string
	sep    string
}

func (f recordLines) write(record recordSource, repeat int, _ string, out *bufio.Writer) error {
	if f.open != "" {
		out.WriteString(f.open)
		out.WriteByte('\n')
	}
	for i := 0; i < repeat; i++ {
		r, err := record()
		if err != nil {
			return err
		}
		if i == 0 && f.header != nil {
			out.WriteString(f.header(r))
			out.WriteByte('\n')
		}
		if i > 0 {
			out.WriteString(f.sep)
		}
		out.WriteString(f.line(r))
	}
	if f.close != "" {
		out.WriteByte('\n')
		out.WriteString(f.close)
	}
	out.WriteByte('\n')
	return out.Flush()
}

// writeSQL streams the records as INSERTs of up to sqlStatementRows rows, one statement per line.
func writeSQL(record recordSource, repeat int, table string, out *bufio.Writer) error {
	for i := 0; i < repeat; i++ {
		r, err := record()
		if err != nil {
			return err
		}
		switch {
		case i == 0:
			out.WriteString(r.SQLInsertHeader(table) + " ")
		case i%sqlStatementRows == 0:
			out.WriteString(";\n" + r.SQLInsertHeader(table) + " ")
		default:
			out.WriteByte(',')
		}
		out.WriteString(r.SQLValues())
	}
	out.WriteString(";\n")
	return out.Flush()
}

// writePGCopy streams the records as one COPY block: its statement, a row each and the \. that ends it.
func writePGCopy(record recordSource, repeat int, table string, out *bufio.Writer) error {
	for i := 0; i < repeat; i++ {
		r, err := record()
		if err != nil {
			return err
		}
		if i == 0 {
			out.WriteString(r.PGCopyHeader(table) + "\n")
		}
		out.WriteString(r.PGCopyLine() + "\n")
	}
	out.WriteString("\\.\n")
	return out.Flush()
}

// writeNoRecord still draws one record: a header's columns come from it, and a lone reference to
// nothing fails as it would with records.
func writeNoRecord(record recordSource, format recordFormat, w io.Writer) error {
	r, err := record()
	if err != nil || format.empty == nil {
		return err
	}
	_, err = io.WriteString(w, format.empty(r))
	return err
}

// defaultTable names the INSERT target when --table is absent: the path's last
// segment, or "records" when stdin is not a lone reference.
func defaultTable(path string) string {
	if path == "" {
		return "records"
	}
	segments := strings.Split(withoutSelectors(path), ".")
	return segments[len(segments)-1]
}

// withoutSelectors is path with its [selectors] cut out.
func withoutSelectors(path string) string {
	var names strings.Builder
	for depth, i := 0, 0; i < len(path); i++ {
		switch {
		case path[i] == '[':
			depth++
		case path[i] == ']':
			depth--
		case depth == 0:
			names.WriteByte(path[i])
		}
	}
	return names.String()
}
