package scanner

import (
	"context"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	os.Setenv("TESTVAR", "ok")
	sc := Scanner{
		PreCommand: []string{"echo", "pre"},
		Command:    []string{"sh", "-c", "echo $TESTVAR"},
		EnvVars:    []string{"TESTVAR"},
	}
	if err := sc.Run(context.Background(), "."); err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
