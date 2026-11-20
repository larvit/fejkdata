package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

type codingAgentLine struct {
	command string
	check   func(t *testing.T, code int, out, errb string)
}

const (
	missingSlashLine = "echo '{sv_SE.personnummer}' | fejkdata"
	seededLine       = "echo '{/sv_SE.personnummer}' | fejkdata --seed 1 --today 2026-11-21"
)

// codingAgentLines holds each command the README's "For coding agents" section shows,
// in the order it shows them.
var codingAgentLines = []codingAgentLine{
	{"fejkdata --list | grep personnummer", func(t *testing.T, code int, out, _ string) {
		if code != 0 || !strings.Contains("\n"+out, "\nsv_SE.personnummer\n") {
			t.Errorf("= %d, %q, want sv_SE.personnummer among the paths", code, out)
		}
	}},
	{missingSlashLine, func(t *testing.T, code int, _, errb string) {
		fixed, out, _ := runAgentLine(strings.Replace(missingSlashLine, "{sv", "{/sv", 1))
		if code != 2 || !strings.Contains(errb, "write {/sv_SE.personnummer}") || fixed != 0 || out == "" {
			t.Errorf("= %d, stderr %q, want exit 2 naming the fix, write {/sv_SE.personnummer}, and the fixed template rendering (%d, %q)", code, errb, fixed, out)
		}
	}},
	{"echo '{/sv_SE.personnummer}' | fejkdata", func(t *testing.T, code int, out, errb string) {
		if code != 0 || !regexp.MustCompile(`^\d{6,8}[-+]\d{4}\n$`).MatchString(out) {
			t.Errorf("= %d, %q, stderr %q, want one personnummer line", code, out, errb)
		}
	}},
	{seededLine, func(t *testing.T, code int, out, errb string) {
		_, again, _ := runAgentLine(seededLine)
		_, otherSeed, _ := runAgentLine(strings.Replace(seededLine, "--seed 1", "--seed 2", 1))
		_, otherDay, _ := runAgentLine(strings.Replace(seededLine, "2026-11-21", "2020-01-01", 1))
		if code != 0 || out == "" || again != out || otherSeed == out || otherDay == out {
			t.Errorf("= %d, %q, stderr %q, again %q, seed 2 %q, another day %q, want the same line again, and another under another seed or day", code, out, errb, again, otherSeed, otherDay)
		}
	}},
}

// runAgentLine runs one README line: an optional echo into fejkdata, then the lines of its output holding grep's word, if the line ends in one.
func runAgentLine(line string) (code int, out, errb string) {
	stages := strings.Split(line, " | ")
	stdin := ""
	if template, ok := strings.CutPrefix(stages[0], "echo "); ok {
		stdin, stages = strings.Trim(template, "'")+"\n", stages[1:]
	}
	code, out, errb = runOut(stdin, strings.Fields(strings.TrimPrefix(stages[0], "fejkdata"))...)
	if len(stages) == 2 {
		var kept []string
		for _, l := range strings.SplitAfter(out, "\n") {
			if strings.Contains(l, strings.TrimPrefix(stages[1], "grep ")) {
				kept = append(kept, l)
			}
		}
		out = strings.Join(kept, "")
	}
	return code, out, errb
}

func readmeCodingAgents(t *testing.T) string {
	t.Helper()
	readme, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, section, ok := strings.Cut(string(readme), "\n### For coding agents\n")
	if !ok {
		t.Fatal("README lost its For coding agents section")
	}
	section, _, _ = strings.Cut(section, "\n### ")
	return section
}

func TestReadmeForCodingAgentsShowsExactlyTheLinesTested(t *testing.T) {
	section := readmeCodingAgents(t)
	var shown []string
	for _, block := range regexp.MustCompile("(?s)```sh\n(.*?)```").FindAllStringSubmatch(section, -1) {
		for _, line := range strings.Split(strings.TrimSpace(block[1]), "\n") {
			command, _, _ := strings.Cut(line, "  #")
			shown = append(shown, strings.TrimSpace(command))
		}
	}
	var tested []string
	for _, l := range codingAgentLines {
		tested = append(tested, l.command)
	}
	if !slices.Equal(shown, tested) {
		t.Errorf("README's For coding agents shows %q, want the tested lines %q", shown, tested)
	}
}

func TestReadmeForCodingAgentsLinesRun(t *testing.T) {
	for _, l := range codingAgentLines {
		t.Run(l.command, func(t *testing.T) {
			code, out, errb := runAgentLine(l.command)
			l.check(t, code, out, errb)
		})
	}
}

func TestReadmeForCodingAgentsSnippetRunsOnlyTestedLines(t *testing.T) {
	_, snippet, ok := strings.Cut(readmeCodingAgents(t), "```markdown\n")
	if !ok {
		t.Fatal("README's For coding agents lost its AGENTS.md snippet")
	}
	snippet, _, _ = strings.Cut(snippet, "```")
	spans := regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(snippet, -1)
	commands := 0
	for _, m := range spans {
		if !strings.HasPrefix(m[1], "fejkdata ") && !strings.HasPrefix(m[1], "echo ") {
			continue
		}
		commands++
		if !slices.ContainsFunc(codingAgentLines, func(l codingAgentLine) bool {
			return l.command == m[1]
		}) {
			t.Errorf("the AGENTS.md snippet runs %q, which no test runs", m[1])
		}
	}
	if commands == 0 {
		t.Error("the AGENTS.md snippet names no command")
	}
}
