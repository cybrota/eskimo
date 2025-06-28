package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github-scanner/internal/config"
	internalgithub "github-scanner/internal/github"
	"github-scanner/internal/scanner"
)

func main() {
	var org string
	var configPath string
	flag.StringVar(&org, "org", "", "GitHub organization")
	flag.StringVar(&configPath, "config", "scanners.yaml", "Scanner config file")
	flag.Parse()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN must be set")
	}
	if org == "" {
		log.Fatal("-org must be specified")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	gh := internalgithub.NewClient(token, org)
	ctx := context.Background()
	repos, err := gh.ListRepos(ctx)
	if err != nil {
		log.Fatalf("listing repos: %v", err)
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
}
