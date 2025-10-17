package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	github "github.com/google/go-github/v55/github"

	"github.com/spf13/cobra"

	"github.com/cybrota/eskimo/internal/auth"
	"github.com/cybrota/eskimo/internal/config"
	internalgithub "github.com/cybrota/eskimo/internal/github"
	"github.com/cybrota/eskimo/internal/scanner"
)

type scanLog struct {
	repo    string
	scanner string
	output  string
	err     error
}

const (
	sampleLimit      = 10
	defaultClonePath = "/tmp/github-repos"
)

var (
	org        string
	configPath string
	sample     bool
	clonePath  string
)

var rootCmd = &cobra.Command{
	Use:   "eskimo",
	Short: "Pluggable security scanner",
	RunE: func(cmd *cobra.Command, args []string) error {
		token := auth.LoadToken()
		if token == "" {
			return fmt.Errorf("GITHUB_TOKEN must be set or run 'eskimo auth'")
		}
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		gh := internalgithub.NewClient(token, org)
		ctx := context.Background()
		repos, err := gh.ListRepos(ctx)
		if err != nil {
			return err
		}
		baseDir, err := sanitizeClonePath(clonePath)
		if err != nil {
			return err
		}
		if err := ensureCloneBase(baseDir); err != nil {
			return err
		}
		fmt.Printf("using clone path %s\n", baseDir)
		totalRepos := len(repos)
		fmt.Printf("found %d repositories\n", totalRepos)
		if sample {
			if totalRepos == 0 {
				fmt.Println("no repositories available to sample")
				return nil
			}
			limit := sampleLimit
			if totalRepos < sampleLimit {
				limit = totalRepos
			}
			repos = repos[:limit]
			fmt.Printf("sampling %d repositories (--sample flag)\n", len(repos))
		}

		// For I/O-bound workloads, use higher parallelism than CPU count
		parallel := runtime.NumCPU() * 4
		sem := make(chan struct{}, parallel)
		type repoInfo struct {
			name string
			path string
		}
		repoCh := make(chan repoInfo, len(repos))
		var cloneWG sync.WaitGroup
		for _, repo := range repos {
			cloneWG.Add(1)
			sem <- struct{}{}
			go func(r *github.Repository) {
				defer cloneWG.Done()
				fmt.Printf("cloning %s...\n", r.GetName())
				repoPath := filepath.Join(baseDir, r.GetName())
				if err := removeExistingRepo(repoPath); err != nil {
					log.Printf("failed to prepare %s: %v", repoPath, err)
					<-sem
					return
				} else {
					log.Printf("Cleaned a similar named repository: %s", *r.Name)
				}
				repoPath, err := gh.CloneRepo(r, baseDir)
				if err != nil {
					log.Printf("failed to clone %s: %v", r.GetName(), err)
					<-sem
					return
				}
				repoCh <- repoInfo{name: r.GetName(), path: repoPath}
				<-sem
			}(repo)
		}
		cloneWG.Wait()
		close(repoCh)

		logCh := make(chan scanLog, len(repos)*len(cfg.Scanners))
		var logWG sync.WaitGroup
		logWG.Add(1)
		go func() {
			defer logWG.Done()
			for l := range logCh {
				prefix := fmt.Sprintf("%s: %s", l.repo, l.scanner)
				if l.err != nil {
					fmt.Printf("%s failed: %v\n%s\n", prefix, l.err, l.output)
				} else if l.output != "" {
					fmt.Printf("%s output:\n%s\n", prefix, l.output)
				}
			}
		}()

		scanSem := make(chan struct{}, parallel)
		var scanWG sync.WaitGroup
		for info := range repoCh {
			scanWG.Add(1)
			scanSem <- struct{}{}
			go func(in repoInfo) {
				defer scanWG.Done()
				var wg sync.WaitGroup
				for _, sc := range cfg.Scanners {
					scCopy := sc
					wg.Add(1)
					go func() {
						defer wg.Done()
						s := scanner.Scanner(scCopy)
						fmt.Printf("%s: running %s...\n", in.name, scCopy.Name)
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

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&org, "org", "", "GitHub organization")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "scanners.yaml", "Scanner config file")
	rootCmd.PersistentFlags().BoolVar(&sample, "sample", false, fmt.Sprintf("run scans on at most %d repositories", sampleLimit))
	rootCmd.PersistentFlags().StringVar(&clonePath, "clone-path", defaultClonePath, "directory used to store cloned repositories")
	rootCmd.MarkPersistentFlagRequired("org")
	rootCmd.AddCommand(authCmd)
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

func ensureCloneBase(baseDir string) error {
	if _, err := os.Lstat(baseDir); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(baseDir, 0755)
		}
		return fmt.Errorf("stat clone path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		return fmt.Errorf("resolve clone path symlink: %w", err)
	}
	if isRootPath(resolved) {
		return fmt.Errorf("clone path resolves to filesystem root (%s) which is not allowed", resolved)
	}
	return nil
}

func removeExistingRepo(repoPath string) error {
	info, err := os.Lstat(repoPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat repo path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(repoPath)
	if err != nil {
		return fmt.Errorf("resolve repo path: %w", err)
	}
	if isRootPath(resolved) {
		return fmt.Errorf("refusing to remove repository path that resolves to root (%s)", resolved)
	}
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("remove existing repo %s: %w", repoPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("repository path %s is a symlink, refusing to remove", repoPath)
	}
	return nil
}
