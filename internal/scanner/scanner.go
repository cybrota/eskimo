package scanner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// Scanner defines a pluggable scanner

type Scanner struct {
	Name       string
	PreCommand []string
	Command    []string
	EnvVars    []string
	Disable    bool
}

func (s Scanner) Run(ctx context.Context, repoPath string) ([]byte, error) {
	env := os.Environ()
	for _, key := range s.EnvVars {
		val := os.Getenv(key)
		env = append(env, fmt.Sprintf("%s=%s", key, val))
	}
	if len(s.PreCommand) > 0 {
		pre := exec.CommandContext(ctx, s.PreCommand[0], s.PreCommand[1:]...)
		pre.Dir = repoPath
		pre.Env = env
		out, err := pre.CombinedOutput()
		if err != nil {
			return out, fmt.Errorf("pre-command failed: %v: %s", err, string(out))
		}
	}
	if len(s.Command) == 0 {
		return nil, fmt.Errorf("no command specified")
	}
	cmd := exec.CommandContext(ctx, s.Command[0], s.Command[1:]...)
	cmd.Env = env
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	return out, err
}
