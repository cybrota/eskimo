package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDeviceFlow(t *testing.T) {
	deviceCh := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept header not set")
		}
		switch r.URL.Path {
		case "/login/device/code":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"device_code":"code","user_code":"u","verification_uri":"https://example.com","interval":1}`))
		case "/login/oauth/access_token":
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "device_code=code") {
				if len(deviceCh) == 0 {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"error":"authorization_pending"}`))
					deviceCh <- struct{}{}
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"access_token":"tok"}`))
				}
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	deviceEndpoint = server.URL + "/login/device/code"
	tokenEndpoint = server.URL + "/login/oauth/access_token"

	opened := false
	openFunc := func(url string) error { opened = true; return nil }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tok, err := DeviceFlow(ctx, "client", "repo", openFunc)
	if err != nil {
		t.Fatalf("flow failed: %v", err)
	}
	if tok != "tok" {
		t.Fatalf("unexpected token: %s", tok)
	}
	if !opened {
		t.Fatalf("browser not opened")
	}
}

func TestSaveLoadToken(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	path, err := SaveToken("secret")
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if path != filepath.Join(tmp, ".config", "eskimo", "token") {
		t.Fatalf("unexpected path: %s", path)
	}
	tok := LoadToken()
	if tok != "secret" {
		t.Fatalf("expected token, got %s", tok)
	}
}
