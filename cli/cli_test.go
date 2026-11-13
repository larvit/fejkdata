package cli

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/larvit/fejkdata"
)

// shipped holds the shipped modules as read from the checkout, each folder under ../data
// holding a manifest, so these tests link no data package.
var shipped = func() []fs.FS {
	var modules []fs.FS
	err := filepath.WalkDir(filepath.Join("..", "data"), func(path string, d fs.DirEntry, err error) error {
		if err == nil && d.Name() == ".fejkdata.json" {
			modules = append(modules, os.DirFS(filepath.Dir(path)))
		}
		return err
	})
	if err != nil || len(modules) == 0 {
		panic(fmt.Sprintf("no shipped module found under ../data: %v", err))
	}
	return modules
}()

// svSE, enUS and misc are copies of a shipped module's files without its manifest, so a test
// layers its categories unprefixed with -d.
var svSE, enUS, misc string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "fejkdata-cli-test")
	if err != nil {
		panic(err)
	}
	svSE, enUS, misc = bareCopy(tmp, "sv_SE"), bareCopy(tmp, "en_US"), bareCopy(tmp, "misc")
	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// bareCopy copies the category and rows files of the shipped module in data/<module> into a
// folder under tmp, and returns that folder.
func bareCopy(tmp, module string) string {
	src, dst := filepath.Join("..", "data", module), filepath.Join(tmp, module)
	if err := os.Mkdir(dst, 0o755); err != nil {
		panic(err)
	}
	for _, pattern := range []string{"*.json", "*.tsv"} {
		files, err := filepath.Glob(filepath.Join(src, pattern))
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if strings.HasPrefix(filepath.Base(f), ".") {
				continue
			}
			b, err := os.ReadFile(f)
			if err == nil {
				err = os.WriteFile(filepath.Join(dst, filepath.Base(f)), b, 0o644)
			}
			if err != nil {
				panic(err)
			}
		}
	}
	return dst
}

func runOut(stdin string, args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run("fejkdata", shipped, args, system{1, testDay}, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

// runTerminal runs as from a shell with nothing piped in.
func runTerminal(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run("fejkdata", shipped, args, system{1, testDay}, nil, &out, &errb)
	return code, out.String(), errb.String()
}

var testDay = time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

var temporaryFolder = regexp.MustCompile(regexp.QuoteMeta(os.TempDir()) + `\S*/`)

func argsName(args []string) string {
	if len(args) == 0 {
		return "no flags"
	}
	return temporaryFolder.ReplaceAllString(strings.Join(args, " "), "")
}

func writeDataFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestRunLoadsTheModulesItIsGiven(t *testing.T) {
	module := fstest.MapFS{"pet.json": {Data: []byte(`{"format":"{name}","name":["Fido"]}`)}}
	var out, errb bytes.Buffer
	if code := run("fejkdata", []fs.FS{module}, nil, system{1, testDay}, strings.NewReader("{/pet}"), &out, &errb); code != 0 || out.String() != "Fido" {
		t.Errorf("run over a module holding pet = %d, %q, %q, want Fido", code, out.String(), errb.String())
	}
	out.Reset()
	if code := run("fejkdata", []fs.FS{module}, []string{"--list"}, system{1, testDay}, nil, &out, &errb); code != 0 || out.String() != "pet\npet.name\n" {
		t.Errorf("--list over a module holding pet = %d, %q, want pet and pet.name only", code, out.String())
	}
}

func TestRunOutputsValue(t *testing.T) {
	code, out, errb := runOut("{/person}", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if out == "" || strings.Contains(out, "\n") {
		t.Fatalf("output = %q, stderr=%q; want one value and nothing added", out, errb)
	}
}

func TestRunDotPath(t *testing.T) {
	code, full, _ := runOut("{/person}", "--seed", "7", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("person run = %d", code)
	}
	code, last, errb := runOut("{/person.last}", "--seed", "7", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("person.last run = %d, stderr=%q", code, errb)
	}
	if strings.TrimSpace(last) == "" {
		t.Fatal("empty person.last output")
	}
	if last == full {
		t.Errorf("person.last %q should differ from person %q", last, full)
	}
}

func TestRunWithoutASeedTakesTheSystemSeed(t *testing.T) {
	var drawn, seeded, errb bytes.Buffer
	if code := run("fejkdata", shipped, nil, system{42, testDay}, strings.NewReader("{uuid()}"), &drawn, &errb); code != 0 {
		t.Fatalf("run with system seed 42 = %d, stderr=%q", code, errb.String())
	}
	if code := run("fejkdata", shipped, []string{"--seed", "42"}, system{7, testDay}, strings.NewReader("{uuid()}"), &seeded, &errb); code != 0 {
		t.Fatalf("run --seed 42 = %d, stderr=%q", code, errb.String())
	}
	if drawn.String() != seeded.String() {
		t.Fatalf("run with system seed 42 = %q, run --seed 42 = %q; want the same draw", drawn.String(), seeded.String())
	}
}

func TestRunWithoutTodayTakesTheSystemDate(t *testing.T) {
	var out, errb bytes.Buffer
	day := time.Date(2031, 2, 3, 0, 0, 0, 0, time.UTC)
	if code := run("fejkdata", shipped, nil, system{1, day}, strings.NewReader("{born(0,0,'2006')}"), &out, &errb); code != 0 || out.String() != "2031" {
		t.Fatalf("run with system date 2031-02-03 = %d, %q, %q; want 2031", code, out.String(), errb.String())
	}
}

func TestRunRepeat(t *testing.T) {
	code, out, errb := runOut("{/word}\n", "--seed", "1", "--repeat", "3", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("repeat=3 gave %d lines: %q", len(lines), out)
	}
	_, short, _ := runOut("{/word}\n", "--seed", "1", "-n", "3", "-d", svSE)
	if short != out {
		t.Errorf("-n 3 = %q, want the same as --repeat 3 %q", short, out)
	}
}

func TestRunSeparator(t *testing.T) {
	code, out, errb := runOut("{/word}", "--repeat", "3", "--separator", ",", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if n := strings.Count(out, "\n"); n != 0 {
		t.Errorf("want no newline, got %d: %q", n, out)
	}
	if !strings.Contains(out, ",") {
		t.Errorf("values should be comma-joined: %q", out)
	}
}

func TestRunRepeatAdvancesRNG(t *testing.T) {
	code, out, errb := runOut("{/person}\n", "--seed", "1", "--repeat", "5", "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	slices.Sort(lines)
	if len(slices.Compact(lines)) < 2 {
		t.Errorf("repeat should vary output, all identical: %q", out)
	}
}

func TestRunMisuse(t *testing.T) {
	for _, c := range []struct {
		stdin string
		args  []string
	}{
		{"{/person}", []string{"-d", svSE, "word"}},
		{"", []string{"-d", svSE, "--list", "person"}},
		{"{/sv_SE.person}", []string{"--no-shipped-data"}},
	} {
		t.Run(fmt.Sprintf("%q %s", c.stdin, argsName(c.args)), func(t *testing.T) {
			code, out, errb := runOut(c.stdin, c.args...)
			if code != 2 {
				t.Errorf("run(%q, %v) = %d, want 2", c.stdin, c.args, code)
			}
			if !strings.Contains(errb, "try 'fejkdata --help'") {
				t.Errorf("run(%q, %v) stderr = %q, want a pointer to --help", c.stdin, c.args, errb)
			}
			if out != "" {
				t.Errorf("run(%q, %v) stdout = %q, want nothing", c.stdin, c.args, out)
			}
		})
	}
}

func TestRunListHelpAndVersionReadNoStdin(t *testing.T) {
	for _, args := range [][]string{{"--list"}, {"--help"}, {"--version"}} {
		t.Run(argsName(args), func(t *testing.T) {
			if code, out, errb := runTerminal(args...); code != 0 || out == "" {
				t.Errorf("run(%v) with nothing piped = %d, %q, %q; want it to run, reading no stdin", args, code, out, errb)
			}
		})
	}
}

func TestRunRendersExactlyWhatIsSent(t *testing.T) {
	for stdin, want := range map[string]string{
		"hihi":                           "hihi",
		"hihi\n":                         "hihi\n",
		"hihi\r\n":                       "hihi\r\n",
		"hihi\n\n":                       "hihi\n\n",
		"{{x}}\n":                        "{x}\n",
		"\"a\"\n":                        "a",
		" \t\"a\"\r\n\n":                 "a",
		"{/misc.territory[SE].alpha2}":   "SE",
		"{/misc.territory[SE].alpha2}\n": "SE\n",
	} {
		t.Run(fmt.Sprintf("%q", stdin), func(t *testing.T) {
			code, out, errb := runOut(stdin, "--seed", "1")
			if code != 0 || out != want {
				t.Errorf("run(%q) = %d, %q, %q; want %q", stdin, code, out, errb, want)
			}
		})
	}
}

func TestRunAddsNoNewlineToAJSONTemplatesFormat(t *testing.T) {
	if code, out, errb := runOut("{\"format\":\"{x}\\n\",\"x\":[\"a\",\"b\"]}\n"); code != 0 || (out != "a\n" && out != "b\n") {
		t.Errorf("a JSON template's format = %d, %q, %q; want its own newline only", code, out, errb)
	}
}

func TestRunWritesTheSeparatorAfterEachRendersOwnNewline(t *testing.T) {
	if code, out, errb := runOut("{/sv_SE.word}\n", "--seed", "1", "-n", "3", "--separator", ","); code != 0 || strings.Count(out, "\n,") != 2 || !strings.HasSuffix(out, "\n") {
		t.Errorf("echo's newline and a separator = %d, %q, %q; want each render's newline, then the separator", code, out, errb)
	}
}

func TestRunRendersBlankStdin(t *testing.T) {
	for _, stdin := range []string{"", "\n", " \r\n\t"} {
		t.Run(fmt.Sprintf("%q", stdin), func(t *testing.T) {
			if code, out, errb := runOut(stdin); code != 0 || out != stdin {
				t.Errorf("run(%q) = %d, %q, %q; want the blank template rendered as it is", stdin, code, out, errb)
			}
		})
	}
}

func TestRunTextPrintsAsWritten(t *testing.T) {
	for _, stdin := range []string{"sv_SE.person", "[Skåne län]", "x[1]y", `say "hi"`} {
		t.Run(stdin, func(t *testing.T) {
			code, out, errb := runOut(stdin)
			if code != 0 || out != stdin {
				t.Errorf("run(%q) = %d, %q, %q; want the text as written", stdin, code, out, errb)
			}
		})
	}
}

func TestRunMultipleDataPaths(t *testing.T) {
	code, out, errb := runOut("{/person}", "-d", enUS, "--data-path", svSE)
	if code != 0 {
		t.Fatalf("run(multi-dir) = %d, stderr=%q", code, errb)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty output for multi-dir run")
	}
}

func TestRunUnknownCategoryFails(t *testing.T) {
	code, _, errb := runOut("{/nope}", "--data-path", svSE, "--format", "csv")
	if code != 1 {
		t.Fatalf("run = %d, want 1", code)
	}
	if !strings.Contains(errb, "nope") {
		t.Errorf("stderr %q should name the unknown category", errb)
	}
}

func TestRunList(t *testing.T) {
	code, out, errb := runOut("", "--data-path", svSE, "--list")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	for _, want := range []string{"person", "person.last", "address"} {
		t.Run(want, func(t *testing.T) {
			if !strings.Contains(out, want) {
				t.Errorf("list output missing %q:\n%s", want, out)
			}
		})
	}
}

func TestRunHelpOnStdout(t *testing.T) {
	for _, h := range []string{"-h", "--help"} {
		t.Run(h, func(t *testing.T) {
			code, out, errb := runOut("", h)
			if code != 0 {
				t.Errorf("%s = %d, want 0", h, code)
			}
			if !strings.Contains(out, "Usage") || errb != "" {
				t.Errorf("%s: stdout = %q, stderr = %q, want usage on stdout only", h, out, errb)
			}
		})
	}
}

func TestCommandNameIsTheFileItRunsFrom(t *testing.T) {
	for arg0, want := range map[string]string{
		"/usr/local/bin/mycli": "mycli",
		"mycli.EXE":            "mycli",
		"mycli.exe":            "mycli",
		"my.cli":               "my.cli",
		"":                     "fejkdata",
	} {
		t.Run(fmt.Sprintf("%q", arg0), func(t *testing.T) {
			if got := commandName(arg0); got != want {
				t.Errorf("commandName(%q) = %q, want %q", arg0, got, want)
			}
		})
	}
}

func TestRunNamesItselfByTheCommandName(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run("mycli", shipped, []string{"--nope"}, system{1, testDay}, nil, &out, &errb); code != 2 || errb.String() != "mycli: unknown flag --nope\ntry 'mycli --help'\n" {
		t.Errorf("mycli --nope = %d, %q", code, errb.String())
	}
	errb.Reset()
	if code := run("mycli", shipped, nil, system{1, testDay}, nil, &out, &errb); code != 2 || !strings.Contains(errb.String(), "| mycli; mycli --list prints") {
		t.Errorf("mycli with nothing piped = %d, %q, want a pipe into mycli", code, errb.String())
	}
	errb.Reset()
	if code := run("mycli", []fs.FS{nil, nil}, nil, system{1, testDay}, strings.NewReader("x"), &out, &errb); code != 1 || strings.Count(errb.String(), "mycli: ") != 2 || strings.Contains(errb.String(), "fejkdata") {
		t.Errorf("mycli over two nil modules = %d, %q, want each line under mycli", code, errb.String())
	}
	errb.Reset()
	if code := run("mycli", shipped, []string{"-d", "/nope"}, system{1, testDay}, strings.NewReader("{/sv_SE.person}"), &out, &errb); code != 1 || !strings.HasPrefix(errb.String(), "mycli: ") || strings.Contains(errb.String(), "fejkdata") {
		t.Errorf("mycli -d /nope = %d, %q, want a load error under mycli", code, errb.String())
	}
	if code := run("mycli", shipped, []string{"--help"}, system{1, testDay}, nil, &out, &errb); code != 0 || !strings.HasPrefix(out.String(), "Usage: <template> | mycli [flags]\n") || strings.Contains(strings.ReplaceAll(out.String(), "https://github.com/larvit/fejkdata#", ""), "fejkdata") {
		t.Errorf("mycli --help = %d, %q, want mycli named throughout, save in links to fejkdata's docs", code, out.String())
	}
}

func TestRunMissingDirFails(t *testing.T) {
	code, _, errb := runOut("{/person}", "--data-path", "../data/nope")
	if code != 1 {
		t.Fatalf("run = %d, want 1", code)
	}
	if errb == "" {
		t.Error("want an error message on stderr")
	}
}

func TestRunShippedDataByDefault(t *testing.T) {
	code, out, errb := runOut("{/sv_SE.person}", "--seed", "1")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Fatalf("run = %d, out=%q, stderr=%q", code, out, errb)
	}
	code, list, _ := runOut("", "--list")
	if code != 0 || !strings.Contains(list, "en_US.person\n") || !strings.Contains(list, "misc.uuid\n") {
		t.Errorf("--list without --data-path = %d, %q", code, list)
	}
	code, _, errb = runOut("{/person}")
	if code != 2 || !strings.Contains(errb, "person") {
		t.Errorf("a category outside the shipped tree: code %d, stderr %q", code, errb)
	}
}

func regionDir(t *testing.T) string {
	t.Helper()
	return writeDataFiles(t, map[string]string{
		"region.json": `{"format":"{name}","rows":"region.tsv","key":"code","name":"name"}`,
		"region.tsv":  "code\tname\n01\tStockholms län\n12\tSkåne län\n",
	})
}

func TestRunSelectsATableRow(t *testing.T) {
	dir := regionDir(t)
	if code, out, errb := runOut("{/region[12]}", "--no-shipped-data", "-d", dir); code != 0 || out != "Skåne län" {
		t.Fatalf("{/region[12]} = %d, %q, stderr=%q", code, out, errb)
	}
	if code, out, _ := runOut("{/region[Skåne län].code}", "--no-shipped-data", "-d", dir); code != 0 || out != "12" {
		t.Fatalf("{/region[Skåne län].code} = %d, %q", code, out)
	}
	if code, out, _ := runOut("{/misc.territory[SE]}", "--seed", "1", "--format", "csv"); code != 0 || !strings.HasPrefix(out, "alpha2,") || !strings.Contains(out, "\nSE,SWE,") {
		t.Fatalf("--format csv {/misc.territory[SE]} = %d, %q", code, out)
	}
}

func TestRunSelectingARowATableLacksIsMisuse(t *testing.T) {
	dir := regionDir(t)
	if code, _, errb := runOut("{/region[99]}", "--no-shipped-data", "-d", dir); code != 2 || !strings.Contains(errb, `"99"`) {
		t.Fatalf("{/region[99]} = %d, stderr=%q, want misuse naming the row", code, errb)
	}
}

func firstSubmatches(matches [][]string) []string {
	firsts := make([]string, len(matches))
	for i, m := range matches {
		firsts[i] = m[1]
	}
	return firsts
}

func distinctSorted(s []string) []string {
	s = slices.Clone(s)
	slices.Sort(s)
	return slices.Compact(s)
}

func TestUsageExamplesRenderWithNoData(t *testing.T) {
	examples := regexp.MustCompile(`(?m)^  echo '([^']*)' \| COMMAND$`).FindAllStringSubmatch(usage+"\n  "+pipeHint("COMMAND")+"\n", -1)
	if len(examples) < 3 {
		t.Fatalf("usage and pipe hint hold %d echo examples, want the help's two and the hint's", len(examples))
	}
	for _, example := range distinctSorted(firstSubmatches(examples)) {
		t.Run(example, func(t *testing.T) {
			if code, out, errb := runOut(example, "--no-shipped-data"); code != 0 || strings.TrimSpace(out) == "" {
				t.Errorf("echo %q with no data: run = %d, %q, stderr %q", example, code, out, errb)
			}
		})
	}
}

func TestUsageNamesNoPathACustomCLIMayNotCarry(t *testing.T) {
	for _, token := range distinctSorted(regexp.MustCompile(`\{/[^}]*\}`).FindAllString(usage, -1)) {
		t.Run(token, func(t *testing.T) {
			if token != "{/path}" && token != "{/table[KEY]}" && token != "{/path as p}" {
				t.Errorf("usage names %s, a path a custom CLI may not carry", token)
			}
		})
	}
}

func TestUsageNamesEveryExitCode(t *testing.T) {
	for _, code := range []string{"0 success", "1 runtime error", "2 misuse"} {
		t.Run(code, func(t *testing.T) {
			if !strings.Contains(usage, code) {
				t.Errorf("usage does not name exit code %s", code)
			}
		})
	}
}

func TestRunNoShippedData(t *testing.T) {
	code, list, errb := runOut("", "--no-shipped-data", "-d", misc, "--list")
	if code != 0 {
		t.Fatalf("run = %d, stderr=%q", code, errb)
	}
	if strings.Contains(list, "sv_SE") || !strings.Contains(list, "uuid\n") {
		t.Errorf("--no-shipped-data --list = %q, want only the given dir", list)
	}
	code, out, _ := runOut("{/uuid}", "--no-shipped-data", "-d", misc, "-s", "3")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Errorf("run = %d, out=%q", code, out)
	}
	if code, out, errb := runOut("{digits(3)}", "--no-shipped-data"); code != 0 || len(out) != 3 {
		t.Errorf("--no-shipped-data alone = %d, %q, %q, want a template reading no data rendered", code, out, errb)
	}
	if code, _, errb := runOut("{/sv_SE.person}", "--no-shipped-data"); code != 2 || !strings.HasSuffix(errb, "no data is loaded\ntry 'fejkdata --help'\n") {
		t.Errorf("--no-shipped-data alone = %d, %q, want a reference to nothing refused", code, errb)
	}
}

func TestRunRepeatOfZeroStillRefusesAReferenceToNothing(t *testing.T) {
	dir := recordDir(t)
	if code, _, _ := runOut("{/nope}", "-n", "0", "--format", "json", "-d", dir); code != 1 {
		t.Errorf("run(-n 0) of a reference to nothing = %d, want 1", code)
	}
}

func TestRunRepeatIsBounded(t *testing.T) {
	for _, args := range [][]string{
		{"--repeat", "9223372036854775807"},
		{"--repeat", "1048577"},
		{"-n", "99999999999"},
	} {
		t.Run(argsName(args), func(t *testing.T) {
			code, out, errb := runOut("{/sv_SE.word}", args...)
			if code != 2 || out != "" || !strings.Contains(errb, "0..1048576") {
				t.Errorf("run(%v) = %d, %q, %q, want misuse naming the range", args, code, out, errb)
			}
		})
	}
}

func TestRunRepeatAtTheCapStreamsEveryRender(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.json"), []byte(`"x"`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errb := runOut("{/x}", "--no-shipped-data", "-d", dir, "--repeat", "1048576", "--separator", "")
	if code != 0 || len(out) != 1048576 {
		t.Errorf("repeat at the cap = %d, %d bytes, stderr %q; want every render streamed", code, len(out), errb)
	}
}

func TestRunNumbersFollowTheShell(t *testing.T) {
	_, want, _ := runOut("{/sv_SE.word}", "--seed", "7", "--repeat", "3")
	code, got, errb := runOut("{/sv_SE.word}", "--seed", "007", "--repeat", "+3")
	if code != 0 || got != want {
		t.Errorf("run(--seed 007 --repeat +3) = %d, %q, stderr %q; want the same as --seed 7 --repeat 3 %q", code, got, errb, want)
	}
}

func TestADataPathCategoryFailingAtFirstReachExits1(t *testing.T) {
	dir := writeDataFiles(t, map[string]string{
		".fejkdata.json": `{"index": {"bad": {"paths": [""]}}}`,
		"bad.json":       `"{/nope}"`,
	})
	for _, args := range [][]string{{"-d", dir}, {"-d", dir, "--format", "json"}} {
		t.Run(argsName(args), func(t *testing.T) {
			if code, _, errb := runOut("{/bad}", args...); code != 1 || !strings.Contains(errb, "{/nope}") {
				t.Errorf("%v on a category failing to load = %d, %q; want exit 1 naming the reference", args, code, errb)
			}
		})
	}
}

func TestHelpLinksTheDocsItIsGiven(t *testing.T) {
	out := help("mycli", "https://github.com/larvit/fejkdata/blob/v0.4.0/README.md")
	if !strings.Contains(out, "https://github.com/larvit/fejkdata/blob/v0.4.0/README.md#names") {
		t.Errorf("help does not link the v0.4.0 README's Names:\n%s", out)
	}
}

func helpBuiltinLines(t *testing.T) []string {
	t.Helper()
	_, out, _ := runOut("", "--help")
	start := strings.Index(out, "builtins:\n")
	if start < 0 {
		t.Fatalf("--help has no list of builtins:\n%s", out)
	}
	var lines []string
	for _, line := range strings.Split(out[start+len("builtins:\n"):], "\n") {
		if !strings.HasPrefix(line, "  ") {
			break
		}
		lines = append(lines, line)
	}
	return lines
}

func TestHelpListsEveryBuiltin(t *testing.T) {
	listed := strings.Fields(strings.Join(helpBuiltinLines(t), " "))
	if want := fejkdata.Builtins(); !slices.Equal(listed, want) {
		t.Fatalf("--help lists the builtins %v, want %v", listed, want)
	}
}

func TestHelpWrapsItsBuiltinsAt88Columns(t *testing.T) {
	lines := helpBuiltinLines(t)
	if len(lines) < 2 {
		t.Errorf("--help lists the builtins on %d line, want them wrapped", len(lines))
	}
	for _, line := range lines {
		t.Run(strings.TrimSpace(line), func(t *testing.T) {
			if len(line) > 88 {
				t.Errorf("--help lists builtins on a line of %d columns, over 88: %q", len(line), line)
			}
		})
	}
}

func TestHelpLeadsToNamesFunctionsAndTheReadme(t *testing.T) {
	_, out, _ := runOut("", "--help")
	if !regexp.MustCompile(`\{[^}]* as [a-z]+\}`).MatchString(out) {
		t.Errorf("--help shows no {x as y}:\n%s", out)
	}
	for _, link := range []string{"https://github.com/larvit/fejkdata#names", "https://github.com/larvit/fejkdata#functions", "https://github.com/larvit/fejkdata#your-own-data"} {
		t.Run(link, func(t *testing.T) {
			if !strings.Contains(out, link) {
				t.Errorf("--help does not link %s", link)
			}
		})
	}
}

func readmeHeadingAnchors(t *testing.T) map[string]bool {
	t.Helper()
	readme, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	anchors := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^#+ (.+)$`).FindAllStringSubmatch(string(readme), -1) {
		anchors[strings.ReplaceAll(strings.ToLower(m[1]), " ", "-")] = true
	}
	return anchors
}

func TestHelpLinksOnlyReadmeHeadings(t *testing.T) {
	anchors := readmeHeadingAnchors(t)
	_, out, _ := runOut("", "--help")
	for _, m := range regexp.MustCompile(`https://github\.com/larvit/fejkdata#(\S+)`).FindAllStringSubmatch(out, -1) {
		t.Run(m[1], func(t *testing.T) {
			if !anchors[m[1]] {
				t.Errorf("--help links #%s, which no README heading makes", m[1])
			}
		})
	}
}

func readmeLoadIntoADatabase(t *testing.T) (users string, pipelines [][]string) {
	t.Helper()
	readme, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, section, _ := strings.Cut(string(readme), "\n### Load into a database\n")
	section, _, _ = strings.Cut(section, "\n## ")
	block := regexp.MustCompile("(?s)```json\n(.*?)```").FindStringSubmatch(section)
	if block == nil {
		t.Fatal("README's Load into a database lost its users.json")
	}
	for _, line := range strings.Split(section, "\n") {
		if stages := strings.Split(line, " | "); len(stages) == 3 && strings.HasPrefix(stages[0], "echo ") {
			pipelines = append(pipelines, stages)
		}
	}
	return block[1], pipelines
}

func TestReadmeLoadIntoADatabaseLeftSidesRun(t *testing.T) {
	users, pipelines := readmeLoadIntoADatabase(t)
	dir := writeDataFiles(t, map[string]string{"users.json": users})
	starts := map[string]string{"pgcopy": "COPY ", "sql": "INSERT INTO", "csv": "id,name\n", "ndjson": `{"id":`}
	for _, stages := range pipelines {
		t.Run(stages[2], func(t *testing.T) {
			args := strings.Fields(strings.ReplaceAll(strings.TrimPrefix(stages[1], "fejkdata "), "./mydata", dir))
			code, out, errb := runOut("{/users}", args...)
			format := args[slices.Index(args, "--format")+1]
			if code != 0 || !strings.HasPrefix(out, starts[format]) || strings.Count(out, "\n") < 1 {
				t.Errorf("README's %q = %d, %q, stderr %q, want a %s run", stages[1], code, out, errb, format)
			}
		})
	}
}

func TestReadmeLoadIntoADatabaseKeepsALineForEachTool(t *testing.T) {
	_, pipelines := readmeLoadIntoADatabase(t)
	for _, tool := range []string{"psql", "sqlite3", "mysql", "mongoimport", "curl"} {
		t.Run(tool, func(t *testing.T) {
			if !slices.ContainsFunc(pipelines, func(stages []string) bool { return strings.Fields(stages[2])[0] == tool }) {
				t.Errorf("README's Load into a database lost its %s line", tool)
			}
		})
	}
}
