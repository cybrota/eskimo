package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	github "github.com/google/go-github/v55/github"

	"github.com/spf13/cobra"

	"github.com/cybrota/eskimo/internal/auth"
	"github.com/cybrota/eskimo/internal/config"
	internalgithub "github.com/cybrota/eskimo/internal/github"
	"github.com/cybrota/eskimo/internal/scanner"
)

var (
	org        string
	configPath string
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
		baseDir := filepath.Join("/tmp", "github-repos")
		os.MkdirAll(baseDir, 0755)
		fmt.Printf("found %d repositories\n", len(repos))

		parallel := runtime.NumCPU()
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
						if err := s.Run(ctx, in.path); err != nil {
							log.Printf("scanner %s failed on %s: %v", scCopy.Name, in.name, err)
						}
					}()
				}
				wg.Wait()
				<-scanSem
			}(info)
		}
		scanWG.Wait()

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&org, "org", "", "GitHub organization")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "scanners.yaml", "Scanner config file")
	rootCmd.MarkPersistentFlagRequired("org")
	rootCmd.AddCommand(authCmd)
}
