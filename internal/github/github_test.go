package github

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	gh "github.com/google/go-github/v55/github"
)

func TestCloneRepo_UpdateExisting(t *testing.T) {
	tmp := t.TempDir()
	repoDir := filepath.Join(tmp, "remote")
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	// init git repo
	run := func(dir string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run(repoDir, "init")
	// create file and commit
	os.WriteFile(filepath.Join(repoDir, "a.txt"), []byte("hello"), 0644)
	run(repoDir, "add", "a.txt")
	run(repoDir, "commit", "-m", "init")

	base := filepath.Join(tmp, "repos")
	os.Mkdir(base, 0755)
	repo := &gh.Repository{Name: gh.String("remote"), CloneURL: gh.String(repoDir)}
	c := &Client{}
	path, err := c.CloneRepo(repo, base)
	if err != nil {
		t.Fatalf("clone1: %v", err)
	}
	if _, err := os.Stat(filepath.Join(path, "a.txt")); err != nil {
		t.Fatalf("file not cloned: %v", err)
	}
	// modify repo
	os.WriteFile(filepath.Join(repoDir, "b.txt"), []byte("world"), 0644)
	run(repoDir, "add", "b.txt")
	run(repoDir, "commit", "-m", "update")

	// second clone should pull
	path2, err := c.CloneRepo(repo, base)
	if err != nil {
		t.Fatalf("clone2: %v", err)
	}
	if path2 != path {
		t.Fatalf("expected same path")
	}
	if _, err := os.Stat(filepath.Join(path, "b.txt")); err != nil {
		t.Fatalf("file not pulled: %v", err)
	}
}

func TestCloneRepo_ReplaceNonGitDir(t *testing.T) {
	tmp := t.TempDir()
	repoDir := filepath.Join(tmp, "remote")
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	run := func(dir string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run(repoDir, "init")
	os.WriteFile(filepath.Join(repoDir, "a.txt"), []byte("hello"), 0644)
	run(repoDir, "add", "a.txt")
	run(repoDir, "commit", "-m", "init")

	base := filepath.Join(tmp, "repos")
	os.Mkdir(base, 0755)
	dest := filepath.Join(base, "remote")
	os.Mkdir(dest, 0755)
	os.WriteFile(filepath.Join(dest, "junk"), []byte("x"), 0644)

	repo := &gh.Repository{Name: gh.String("remote"), CloneURL: gh.String(repoDir)}
	c := &Client{}
	path, err := c.CloneRepo(repo, base)
	if err != nil {
		t.Fatalf("clone: %v", err)
	}
	if path != dest {
		t.Fatalf("expected %s, got %s", dest, path)
	}
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		t.Fatalf("repo not cloned: %v", err)
	}
}
