package scanner

import (
	"context"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	t.Setenv("TESTVAR", "ok")
	sc := Scanner{
		Command: []string{"sh", "-c", "echo $TESTVAR"},
		EnvVars: []string{"TESTVAR"},
	}
	out, err := sc.Run(context.Background(), ".")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(string(out), "ok") {
		t.Fatalf("expected output to contain 'ok', got %s", string(out))
	}
}

func TestRunPreCommand(t *testing.T) {
	sc := Scanner{
		PreCommand: []string{"sh", "-c", "echo pre"},
	}
	out, err := sc.RunPreCommand(context.Background(), ".")
	if err != nil {
		t.Fatalf("run pre-command failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "pre" {
		t.Fatalf("expected pre-command output to be 'pre', got %q", strings.TrimSpace(string(out)))
	}
}
