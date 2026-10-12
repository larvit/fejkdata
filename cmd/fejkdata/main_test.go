package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("FEJKDATA_RUN_MAIN") == "1" {
		main()
	}
	os.Exit(m.Run())
}

func TestMainRendersTheShippedData(t *testing.T) {
	cmd := exec.Command(os.Args[0], "--seed", "1")
	cmd.Env = append(os.Environ(), "FEJKDATA_RUN_MAIN=1")
	cmd.Stdin = strings.NewReader("{/misc.territory[SE].alpha2} {/en_US.person.last} {/sv_SE.address.locality}")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if fields := strings.Fields(string(out)); err != nil || len(fields) < 3 || fields[0] != "SE" {
		t.Errorf("fejkdata = %v, %q, %q, want SE followed by a last name and a locality", err, out, stderr.String())
	}
}

func TestMainWithoutASeedDrawsANewSeedEachRun(t *testing.T) {
	var outs [2]string
	for i := range outs {
		cmd := exec.Command(os.Args[0])
		cmd.Env = append(os.Environ(), "FEJKDATA_RUN_MAIN=1")
		cmd.Stdin = strings.NewReader("{uuid()}")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil || len(out) == 0 {
			t.Fatalf("fejkdata without --seed = %v, %q, %q", err, out, stderr.String())
		}
		outs[i] = string(out)
	}
	if outs[0] == outs[1] {
		t.Errorf("two runs without --seed both printed %q, want a new seed each run", outs[0])
	}
}
