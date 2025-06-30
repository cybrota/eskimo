package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
		for i, repo := range repos {
			fmt.Printf("[%d/%d] cloning %s...\n", i+1, len(repos), repo.GetName())
			repoPath, err := gh.CloneRepo(repo, baseDir)
			if err != nil {
				log.Printf("failed to clone %s: %v", repo.GetName(), err)
				continue
			}
			for _, sc := range cfg.Scanners {
				s := scanner.Scanner(sc)
				fmt.Printf("  running %s...\n", sc.Name)
				if err := s.Run(ctx, repoPath); err != nil {
					log.Printf("scanner %s failed on %s: %v", sc.Name, repo.GetName(), err)
				}
			}
		}
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
