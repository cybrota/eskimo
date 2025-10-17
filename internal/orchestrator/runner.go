// Copyright (c) 2025 Naren Yellavula & Cybrota contributors
// Apache License, Version 2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	github "github.com/google/go-github/v55/github"

	"github.com/cybrota/eskimo/internal/config"
	internalgithub "github.com/cybrota/eskimo/internal/github"
	"github.com/cybrota/eskimo/internal/scanner"
)

const SampleLimit = 10

type scanLog struct {
	repo    string
	scanner string
	output  string
	err     error
}

type Options struct {
	ClonePath string
	Sample    bool
}

type Runner struct {
	logger *slog.Logger
	client *internalgithub.Client
	cfg    *config.Config
	opts   Options
}

func NewRunner(logger *slog.Logger, client *internalgithub.Client, cfg *config.Config, opts Options) *Runner {
	return &Runner{
		logger: logger,
		client: client,
		cfg:    cfg,
		opts:   opts,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	repos, err := r.client.ListRepos(ctx)
	if err != nil {
		return err
	}
	baseDir, err := sanitizeClonePath(r.opts.ClonePath)
	if err != nil {
		return err
	}
	baseDir, err = ensureCloneBase(baseDir)
	if err != nil {
		return err
	}
	r.logger.Info("using clone path", slog.String("path", baseDir))

	totalRepos := len(repos)
	r.logger.Info("repositories discovered", slog.Int("count", totalRepos))
	if r.opts.Sample {
		if totalRepos == 0 {
			r.logger.Info("no repositories available to sample", slog.Int("count", totalRepos))
			return nil
		}
		limit := SampleLimit
		if totalRepos < SampleLimit {
			limit = totalRepos
		}
		repos = repos[:limit]
		r.logger.Info("sampling repositories", slog.Int("count", len(repos)), slog.Int("limit", SampleLimit))
	}

	parallel := runtime.NumCPU() * 4
	sem := make(chan struct{}, parallel)
	type repoInfo struct {
		name string
		path string
	}
	repoCh := make(chan repoInfo, len(repos))
	clonedRepos := make([]repoInfo, 0, len(repos))
	var cloneWG sync.WaitGroup
	for _, repo := range repos {
		cloneWG.Add(1)
		sem <- struct{}{}
		go func(rp *github.Repository) {
			defer cloneWG.Done()
			repoName := rp.GetName()
			r.logger.Info("preparing repository", slog.String("repo", repoName))
			repoPath := filepath.Join(baseDir, repoName)
			removed, err := removeExistingRepo(repoPath)
			if err != nil {
				r.logger.Error("failed to prepare repository directory", slog.String("repo", repoName), slog.String("path", repoPath), slog.Any("error", err))
				<-sem
				return
			}
			if removed {
				r.logger.Info("removed existing repository directory", slog.String("repo", repoName), slog.String("path", repoPath))
			}
			r.logger.Info("cloning repository", slog.String("repo", repoName), slog.String("path", repoPath))
			repoPath, err = r.client.CloneRepo(rp, baseDir)
			if err != nil {
				r.logger.Error("failed to clone repository", slog.String("repo", repoName), slog.Any("error", err))
				<-sem
				return
			}
			repoCh <- repoInfo{name: repoName, path: repoPath}
			<-sem
		}(repo)
	}
	cloneWG.Wait()
	close(repoCh)

	logCh := make(chan scanLog, len(repos)*len(r.cfg.Scanners))
	var logWG sync.WaitGroup
	logWG.Add(1)
	go func() {
		defer logWG.Done()
		for l := range logCh {
			prefix := fmt.Sprintf("%s: %s", l.repo, l.scanner)
			if l.err != nil {
				if l.output != "" {
					fmt.Fprintf(os.Stderr, "%s failed: %v\n%s", prefix, l.err, l.output)
					if !strings.HasSuffix(l.output, "\n") {
						fmt.Fprintln(os.Stderr)
					}
				} else {
					fmt.Fprintf(os.Stderr, "%s failed: %v\n", prefix, l.err)
				}
			} else if l.output != "" {
				fmt.Printf("%s output:\n%s", prefix, l.output)
				if !strings.HasSuffix(l.output, "\n") {
					fmt.Println()
				}
			}
		}
	}()

	scanSem := make(chan struct{}, parallel)
	var scanWG sync.WaitGroup
	for info := range repoCh {
		clonedRepos = append(clonedRepos, info)
		scanWG.Add(1)
		scanSem <- struct{}{}
		go func(in repoInfo) {
			defer scanWG.Done()
			var wg sync.WaitGroup
			for _, sc := range r.cfg.Scanners {
				scCopy := sc
				wg.Add(1)
				go func() {
					defer wg.Done()
					s := scanner.Scanner(scCopy)
					r.logger.Info("running scanner", slog.String("repo", in.name), slog.String("scanner", scCopy.Name))
					out, err := s.Run(ctx, in.path)
					logCh <- scanLog{repo: in.name, scanner: scCopy.Name, output: string(out), err: err}
				}()
			}
			wg.Wait()
			<-scanSem
		}(info)
	}
	scanWG.Wait()
	close(logCh)
	logWG.Wait()

	for _, repo := range clonedRepos {
		if err := os.RemoveAll(repo.path); err != nil {
			r.logger.Error("failed to clean repository directory", slog.String("repo", repo.name), slog.String("path", repo.path), slog.Any("error", err))
			continue
		}
		r.logger.Info("removed repository directory", slog.String("repo", repo.name), slog.String("path", repo.path))
	}

	if err := os.Remove(baseDir); err != nil && !os.IsNotExist(err) {
		if !errors.Is(err, syscall.ENOTEMPTY) {
			r.logger.Warn("unable to remove clone path directory", slog.String("path", baseDir), slog.Any("error", err))
		}
	}

	r.logger.Info("scanning completed successfully", slog.Int("repositories", len(clonedRepos)))

	return nil
}

func sanitizeClonePath(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("clone path cannot be empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("resolve clone path: %w", err)
	}
	clean := filepath.Clean(abs)
	if clean == "." || isRootPath(clean) {
		return "", fmt.Errorf("clone path %q is not allowed", clean)
	}
	return clean, nil
}

func ensureCloneBase(baseDir string) (string, error) {
	info, err := os.Lstat(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(baseDir, 0755); err != nil {
				return "", fmt.Errorf("create clone path: %w", err)
			}
			return baseDir, nil
		}
		return "", fmt.Errorf("stat clone path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(baseDir)
		if err != nil {
			return "", fmt.Errorf("resolve clone path symlink: %w", err)
		}
		if isRootPath(resolved) {
			return "", fmt.Errorf("clone path resolves to filesystem root (%s) which is not allowed", resolved)
		}
		baseDir = resolved
		info, err = os.Stat(baseDir)
		if err != nil {
			return "", fmt.Errorf("stat resolved clone path: %w", err)
		}
	}
	if !info.IsDir() {
		return "", fmt.Errorf("clone path %s is not a directory", baseDir)
	}
	if isRootPath(baseDir) {
		return "", fmt.Errorf("clone path resolves to filesystem root (%s) which is not allowed", baseDir)
	}
	return baseDir, nil
}

func removeExistingRepo(repoPath string) (bool, error) {
	info, err := os.Lstat(repoPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat repo path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("repository path %s is a symlink, refusing to remove", repoPath)
	}
	resolved, err := filepath.EvalSymlinks(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("resolve repo path: %w", err)
	}
	if isRootPath(resolved) {
		return false, fmt.Errorf("refusing to remove repository path that resolves to root (%s)", resolved)
	}
	if err := os.RemoveAll(repoPath); err != nil {
		return false, fmt.Errorf("remove existing repo %s: %w", repoPath, err)
	}
	return true, nil
}

func isRootPath(path string) bool {
	clean := filepath.Clean(path)
	if clean == "." || clean == string(filepath.Separator) {
		return true
	}
	if runtime.GOOS == "windows" {
		vol := filepath.VolumeName(clean)
		if vol != "" {
			rest := strings.TrimPrefix(clean, vol)
			rest = strings.TrimPrefix(rest, string(filepath.Separator))
			return rest == ""
		}
	}
	return filepath.Dir(clean) == clean
}
