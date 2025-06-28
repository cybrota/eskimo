package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	data := []byte(`scanners:
  - name: test
    command: ["echo", "hello"]
    env: ["TEST_ENV"]
`)
	tmp, err := os.CreateTemp("", "cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	cfg, err := Load(tmp.Name())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(cfg.Scanners) != 1 {
		t.Fatalf("expected 1 scanner, got %d", len(cfg.Scanners))
	}
	sc := cfg.Scanners[0]
	if sc.Name != "test" {
		t.Errorf("unexpected scanner name: %s", sc.Name)
	}
}

func TestLoadDisabled(t *testing.T) {
	data := []byte(`scanners:
  - name: test
    command: ["echo", "hello"]
    disable: true
`)
	tmp, err := os.CreateTemp("", "cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	cfg, err := Load(tmp.Name())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(cfg.Scanners) != 0 {
		t.Fatalf("expected 0 scanners, got %d", len(cfg.Scanners))
	}
}

func TestLoadDisableFalse(t *testing.T) {
	data := []byte(`scanners:
  - name: test
    command: ["echo", "hello"]
    disable: false
`)
	tmp, err := os.CreateTemp("", "cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	cfg, err := Load(tmp.Name())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(cfg.Scanners) != 1 {
		t.Fatalf("expected 1 scanner, got %d", len(cfg.Scanners))
	}
	if cfg.Scanners[0].Disable {
		t.Fatalf("scanner should not be disabled")
	}
}
