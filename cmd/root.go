package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/cybrota/eskimo/internal/auth"
	"github.com/cybrota/eskimo/internal/config"
	internalgithub "github.com/cybrota/eskimo/internal/github"
	"github.com/cybrota/eskimo/internal/orchestrator"
)

const defaultClonePath = "/tmp/github-repos"

var (
	org        string
	configPath string
	sample     bool
	clonePath  string
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

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
		runner := orchestrator.NewRunner(logger, gh, cfg, orchestrator.Options{
			ClonePath: clonePath,
			Sample:    sample,
		})
		return runner.Run(context.Background())
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&org, "org", "", "GitHub organization")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "scanners.yaml", "Scanner config file")
	rootCmd.PersistentFlags().BoolVar(&sample, "sample", false, fmt.Sprintf("run scans on at most %d repositories", orchestrator.SampleLimit))
	rootCmd.PersistentFlags().StringVar(&clonePath, "clone-path", defaultClonePath, "directory used to store cloned repositories")
	rootCmd.MarkPersistentFlagRequired("org")
	rootCmd.AddCommand(authCmd)
}
