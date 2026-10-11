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
