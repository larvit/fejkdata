package cli

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	emitCustomers = `{ "format": "", "id": "{seq()}", "name": ["Ada", "Bo", "Cy", "Di", "Ed"] }`
	emitOrders    = `{ "format": "", "id": "{seq()}", "customer_id": "{/run.customers as c}{c.id}", "customer": "{c.name}" }`
	emitProducts  = `{ "format": "", "sku": "{upper(4)}" }`
)

func emitDir(t *testing.T) string {
	t.Helper()
	return writeDataFiles(t, map[string]string{"customers.json": emitCustomers, "orders.json": emitOrders, "products.json": emitProducts})
}

// runEmit runs as from a shell with nothing piped in, over dir alone.
func runEmit(dir string, args ...string) (int, string, string) {
	return runTerminal(append([]string{"--no-shipped-data", "-d", dir}, args...)...)
}

func csvFile(t *testing.T, path string) [][]string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(string(b))).ReadAll()
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return rows
}

func TestRunEmitWritesParentsFirstWithEachTemplateTheInsertTarget(t *testing.T) {
	code, out, errb := runEmit(emitDir(t), "--format", "sql", "--emit", "orders=3", "--emit", "customers=2")
	if code != 0 {
		t.Fatalf("run = %d, %q", code, errb)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], `INSERT INTO "customers" ("id", "name") VALUES `) ||
		!strings.HasPrefix(lines[1], `INSERT INTO "orders" ("id", "customer_id", "customer") VALUES `) {
		t.Fatalf("out = %q, want customers' INSERT, then orders'", out)
	}
	if got := strings.Count(lines[0], "("); got != 3 {
		t.Errorf("customers' INSERT = %q, want 2 rows", lines[0])
	}
	if got := strings.Count(lines[1], "("); got != 4 {
		t.Errorf("orders' INSERT = %q, want 3 rows", lines[1])
	}
}

func TestRunEmitChildReadsARowItsParentWrote(t *testing.T) {
	dir, out := emitDir(t), t.TempDir()
	code, stdout, errb := runEmit(dir, "--format", "csv", "--out-dir", out, "--emit", "customers=4", "--emit", "orders=40")
	if code != 0 || stdout != "" {
		t.Fatalf("run = %d, stdout %q, stderr %q, want 0 and nothing on stdout", code, stdout, errb)
	}
	customers := csvFile(t, filepath.Join(out, "customers.csv"))
	orders := csvFile(t, filepath.Join(out, "orders.csv"))
	if len(customers) != 5 || len(orders) != 41 {
		t.Fatalf("%d customers lines and %d orders lines, want 5 and 41", len(customers), len(orders))
	}
	nameOf := map[string]string{}
	for _, row := range customers[1:] {
		nameOf[row[0]] = row[1]
	}
	read := map[string]bool{}
	for i, row := range orders[1:] {
		name, exists := nameOf[row[1]]
		if !exists || name != row[2] {
			t.Fatalf("order %v names customer %q %q, want a row of %v", row, row[1], row[2], nameOf)
		}
		if row[0] != strconv.Itoa(i+1) {
			t.Fatalf("order %d has id %q, want its own count from 1", i+1, row[0])
		}
		read[row[1]] = true
	}
	if len(read) < 2 {
		t.Errorf("40 orders read %d customers, want more than one", len(read))
	}
}

func TestRunEmitRefusedAtLoad(t *testing.T) {
	cycle := writeDataFiles(t, map[string]string{
		"a.json": `{ "format": "", "id": "{seq()}", "b": "{/run.b.id}" }`,
		"b.json": `{ "format": "", "id": "{seq()}", "a": "{/run.a.id}" }`,
	})
	reserved := writeDataFiles(t, map[string]string{"customers.json": emitCustomers, "run.json": emitProducts})
	dir := emitDir(t)
	for _, c := range []struct {
		name string
		dir  string
		args []string
		want []string
	}{
		{"a template another reads is not named", dir, []string{"--emit", "orders=3"}, []string{"orders", "customers", "--emit"}},
		{"a parent of count 0 under a child above 0", dir, []string{"--emit", "customers=0", "--emit", "orders=3"}, []string{"orders", "customers", "0"}},
		{"a cycle", cycle, []string{"--emit", "a=1", "--emit", "b=1"}, []string{"a", "b", "cycle"}},
		{"a category named run", reserved, []string{"--emit", "customers=1"}, []string{"run"}},
		{"a template named twice", dir, []string{"--emit", "customers=1", "--emit", "customers=2"}, []string{"customers", "twice"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			code, out, errb := runEmit(c.dir, append([]string{"--format", "sql"}, c.args...)...)
			if code != 2 || out != "" {
				t.Fatalf("run = %d, stdout %q, stderr %q, want 2 and no output", code, out, errb)
			}
			for _, word := range c.want {
				if !strings.Contains(errb, word) {
					t.Errorf("stderr %q should say %q", errb, word)
				}
			}
		})
	}
}

func TestRunEmitBothCountsZeroIsFine(t *testing.T) {
	code, out, errb := runEmit(emitDir(t), "--format", "sql", "--emit", "customers=0", "--emit", "orders=0")
	if code != 0 || out != "" {
		t.Fatalf("run = %d, %q, %q, want 0 and nothing written", code, out, errb)
	}
}

func TestRunEmitCountZeroWritesTheHeader(t *testing.T) {
	out := t.TempDir()
	if code, _, errb := runEmit(emitDir(t), "--format", "csv", "--out-dir", out, "--emit", "customers=0", "--emit", "orders=0"); code != 0 {
		t.Fatalf("run = %d, %q", code, errb)
	}
	for file, want := range map[string]string{"customers.csv": "id,name\n", "orders.csv": "id,customer_id,customer\n"} {
		if b, err := os.ReadFile(filepath.Join(out, file)); err != nil || string(b) != want {
			t.Errorf("%s = %q, %v, want %q", file, b, err, want)
		}
	}
}

func TestRunEmitSeveralTemplatesNeedOutDirUnlessTheFormatNamesItsTarget(t *testing.T) {
	dir := emitDir(t)
	for _, format := range []string{"csv", "tsv", "ndjson", "json"} {
		code, out, errb := runEmit(dir, "--format", format, "--emit", "customers=1", "--emit", "orders=1")
		if code != 2 || out != "" || !strings.Contains(errb, "--out-dir") {
			t.Errorf("--format %s = %d, %q, %q, want 2 naming --out-dir", format, code, out, errb)
		}
	}
	for _, format := range []string{"sql", "pgcopy"} {
		if code, out, errb := runEmit(dir, "--format", format, "--emit", "customers=1", "--emit", "orders=1"); code != 0 || !strings.Contains(out, `"orders"`) {
			t.Errorf("--format %s = %d, %q, %q, want both templates on stdout", format, code, out, errb)
		}
	}
	if code, out, errb := runEmit(dir, "--format", "csv", "--emit", "customers=2"); code != 0 || !strings.HasPrefix(out, "id,name\n1,") {
		t.Errorf("one template as csv = %d, %q, %q, want its rows on stdout", code, out, errb)
	}
}

func TestRunEmitOutDirNamesEachFileByTemplateAndFormat(t *testing.T) {
	out := filepath.Join(t.TempDir(), "made", "here")
	if code, stdout, errb := runEmit(emitDir(t), "--format", "sql", "--out-dir", out, "--emit", "customers=1", "--emit", "orders=1"); code != 0 || stdout != "" {
		t.Fatalf("run = %d, %q, %q", code, stdout, errb)
	}
	for _, file := range []string{"customers.sql", "orders.sql"} {
		if b, err := os.ReadFile(filepath.Join(out, file)); err != nil || !strings.HasPrefix(string(b), "INSERT INTO ") {
			t.Errorf("%s = %q, %v, want an INSERT", file, b, err)
		}
	}
}

func TestRunEmitRefusedBesideOtherFlagsAndStdin(t *testing.T) {
	dir := emitDir(t)
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"--format", "sql", "--emit", "customers=1", "--table", "t"}, "--table"},
		{[]string{"--format", "sql", "--emit", "customers=1", "-n", "3"}, "--repeat"},
		{[]string{"--emit", "customers=1"}, "--format"},
		{[]string{"--format", "sql", "--emit", "customers"}, "customers=<count>"},
		{[]string{"--format", "sql", "--emit", "customers=many"}, "many"},
		{[]string{"--format", "sql", "--emit", "customers=-1"}, "-1"},
		{[]string{"--format", "sql", "--emit", "=3"}, "--emit"},
		{[]string{"--format", "sql", "--out-dir", "x"}, "--emit"},
	} {
		t.Run(argsName(c.args), func(t *testing.T) {
			code, out, errb := runEmit(dir, c.args...)
			if code != 2 || out != "" || !strings.Contains(errb, c.want) {
				t.Errorf("run = %d, %q, %q, want 2 naming %s", code, out, errb, c.want)
			}
		})
	}
	code, out, errb := runOut("{/customers}", "--no-shipped-data", "-d", dir, "--format", "sql", "--emit", "customers=1")
	if code != 2 || out != "" || !strings.Contains(errb, "stdin") {
		t.Errorf("--emit with a template on stdin = %d, %q, %q, want 2 naming stdin", code, out, errb)
	}
	if code, out, errb := runOut("", "--no-shipped-data", "-d", dir, "--format", "sql", "--emit", "customers=1"); code != 0 || out == "" {
		t.Errorf("--emit with an empty stdin = %d, %q, %q, want the rows", code, out, errb)
	}
}

func TestRunEmitAddingATemplateKeepsAnothersRows(t *testing.T) {
	dir := emitDir(t)
	rows := func(template string, args ...string) string {
		t.Helper()
		code, out, errb := runEmit(dir, append([]string{"--seed", "7", "--format", "pgcopy"}, args...)...)
		if code != 0 {
			t.Fatalf("run(%v) = %d, %q", args, code, errb)
		}
		_, after, found := strings.Cut(out, `COPY "`+template+`"`)
		block, _, _ := strings.Cut(after, "\\.\n")
		if !found || block == "" {
			t.Fatalf("run(%v) = %q, want a COPY block of %s", args, out, template)
		}
		return block
	}
	alone := rows("customers", "--emit", "customers=5")
	if beside := rows("customers", "--emit", "customers=5", "--emit", "orders=9", "--emit", "products=2"); beside != alone {
		t.Errorf("customers beside two templates = %q, alone %q, want the same rows", beside, alone)
	}
	two := rows("orders", "--emit", "customers=5", "--emit", "orders=9")
	if three := rows("orders", "--emit", "products=2", "--emit", "customers=5", "--emit", "orders=9"); three != two {
		t.Errorf("orders beside products = %q, without %q, want the same rows", three, two)
	}
}

func TestRunEmitUnrelatedTemplatesComeInFlagOrder(t *testing.T) {
	code, out, errb := runEmit(emitDir(t), "--format", "pgcopy", "--emit", "products=1", "--emit", "orders=1", "--emit", "customers=1")
	if code != 0 {
		t.Fatalf("run = %d, %q", code, errb)
	}
	var targets []string
	for _, line := range strings.Split(out, "\n") {
		if rest, isCopy := strings.CutPrefix(line, `COPY "`); isCopy {
			name, _, _ := strings.Cut(rest, `"`)
			targets = append(targets, name)
		}
	}
	if got := strings.Join(targets, " "); got != "products customers orders" {
		t.Errorf("COPY targets = %q, want products, then customers ahead of its child orders", got)
	}
}

func TestHelpDescribesEmit(t *testing.T) {
	_, out, _ := runTerminal("--help")
	for _, want := range []string{"--emit T=N", "--out-dir D", "{/run."} {
		if !strings.Contains(out, want) {
			t.Errorf("--help should mention %s", want)
		}
	}
}
