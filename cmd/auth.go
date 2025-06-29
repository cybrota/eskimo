package cmd

import (
	"errors"
	"fmt"
	"os"

	"eskimo/internal/auth"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authorize via GitHub device flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		if org == "" {
			return errors.New("--org required")
		}
		clientID := os.Getenv("GITHUB_CLIENT_ID")
		if clientID == "" {
			return fmt.Errorf("GITHUB_CLIENT_ID must be set")
		}

		token, err := auth.DeviceFlow(cmd.Context(), clientID, "repo", auth.DefaultBrowser, cmd.OutOrStdout())
		if err != nil {
			return err
		}
		path, err := auth.SaveToken(token)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Authorization is successful. Token saved to %s\n", path)
		return nil
	},
}

func init() {
	authCmd.Flags().StringVar(&org, "org", "", "GitHub organization")
	authCmd.MarkFlagRequired("org")
}
