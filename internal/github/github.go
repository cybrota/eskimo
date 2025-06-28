package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

// Client wraps GitHub client for fetching repositories

type Client struct {
	org    string
	client *github.Client
	token  string
}

func NewClient(token, org string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		org:    org,
		client: github.NewClient(tc),
		token:  token,
	}
}

func (c *Client) ListRepos(ctx context.Context) ([]*github.Repository, error) {
	var all []*github.Repository
	opt := &github.RepositoryListByOrgOptions{Type: "all", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, c.org, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

func (c *Client) CloneRepo(repo *github.Repository, baseDir string) (string, error) {
	if repo.Name == nil {
		return "", fmt.Errorf("repo name is nil")
	}
	name := *repo.Name
	repoURL := repo.GetCloneURL()
	dest := filepath.Join(baseDir, name)

	if _, err := os.Stat(dest); err == nil {
		cmd := exec.Command("git", "-C", dest, "pull")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git pull failed: %v: %s", err, string(out))
		}
		return dest, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	authURL := repoURL
	if c.token != "" {
		authURL = fmt.Sprintf("https://%s@%s", c.token, repoURL[len("https://"):len(repoURL)])
	}
	cmd := exec.Command("git", "clone", "--depth", "1", authURL, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone failed: %v: %s", err, string(out))
	}
	return dest, nil
}
