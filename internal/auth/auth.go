package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type deviceCodeResp struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

var (
	deviceEndpoint = "https://github.com/login/device/code"
	tokenEndpoint  = "https://github.com/login/oauth/access_token"
)

// DeviceFlow performs GitHub device authorization flow.
// openBrowser is called with the verification URL to open for user login.
func DeviceFlow(ctx context.Context, clientID, scope string, openBrowser func(string) error) (string, error) {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("scope", scope)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deviceEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var dc deviceCodeResp
	if err := json.NewDecoder(resp.Body).Decode(&dc); err != nil {
		return "", err
	}
	verURL := dc.VerificationURIComplete
	if verURL == "" {
		verURL = dc.VerificationURI
	}
	if openBrowser != nil {
		openBrowser(verURL)
	}
	ticker := time.NewTicker(time.Duration(dc.Interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			v := url.Values{}
			v.Set("client_id", clientID)
			v.Set("device_code", dc.DeviceCode)
			v.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(v.Encode()))
			if err != nil {
				return "", err
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "application/json")
			r, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			var tr tokenResp
			if err := json.NewDecoder(r.Body).Decode(&tr); err != nil {
				r.Body.Close()
				return "", err
			}
			r.Body.Close()
			if tr.AccessToken != "" {
				return tr.AccessToken, nil
			}
			switch tr.Error {
			case "authorization_pending":
				continue
			case "slow_down":
				ticker.Reset(time.Duration(dc.Interval+5) * time.Second)
			default:
				return "", fmt.Errorf("authorization failed: %s", tr.Error)
			}
		}
	}
}

// SaveToken stores the token in ~/.config/eskimo/token with 0600 permissions.
func SaveToken(token string) (string, error) {
	dir := filepath.Join(os.Getenv("HOME"), ".config", "eskimo")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "token")
	if err := os.WriteFile(path, []byte(token), 0600); err != nil {
		return "", err
	}
	return path, nil
}

// LoadToken returns GITHUB_TOKEN environment variable or stored token if present.
func LoadToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	path := filepath.Join(os.Getenv("HOME"), ".config", "eskimo", "token")
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// DefaultBrowser attempts to open a URL in the default browser.
func DefaultBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return execCommand(cmd, args...)
}

var execCommand = func(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}
