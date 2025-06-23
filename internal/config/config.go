package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Scanner struct {
	Name       string   `yaml:"name"`
	PreCommand []string `yaml:"pre_command"`
	Command    []string `yaml:"command"`
	EnvVars    []string `yaml:"env"`
}

type Config struct {
	Scanners []Scanner `yaml:"scanners"`
}

func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
