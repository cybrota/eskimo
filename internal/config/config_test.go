package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	data := []byte(`scanners:
  - name: test
    command: ["echo", "hello"]
    env: ["TEST_ENV"]
`)
	tmp, err := ioutil.TempFile("", "cfg-*.yaml")
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
