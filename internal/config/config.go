package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Scanner struct {
	Name       string   `yaml:"name"`
	PreCommand []string `yaml:"pre_command"`
	Command    []string `yaml:"command"`
	EnvVars    []string `yaml:"env"`
	Disable    bool     `yaml:"disable"`
}

type Config struct {
	Scanners []Scanner `yaml:"scanners"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	var active []Scanner
	for _, sc := range cfg.Scanners {
		if sc.Disable {
			continue
		}
		active = append(active, sc)
	}
	cfg.Scanners = active
	return &cfg, nil
}
