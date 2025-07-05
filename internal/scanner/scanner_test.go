package scanner

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	os.Setenv("TESTVAR", "ok")
	sc := Scanner{
		PreCommand: []string{"echo", "pre"},
		Command:    []string{"sh", "-c", "echo $TESTVAR"},
		EnvVars:    []string{"TESTVAR"},
	}
	out, err := sc.Run(context.Background(), ".")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(string(out), "ok") {
		t.Fatalf("expected output to contain 'ok', got %s", string(out))
	}
}
