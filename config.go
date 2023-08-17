package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type FeedConfig struct {
	Module  string         `yaml:"module"`
	Options map[string]any `yaml:"options"`
	Name    string         `yaml:"name"`
}

type Config struct {
	Feeds      []FeedConfig `yaml:"feeds"`
	OutputPath string       `yaml:"output_path"`
}

func setDefaults(config *Config) {
	if config.OutputPath == "" {
		config.OutputPath = "./"
	}
}

func getFeedsFromConfig(path string) (*Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&config)

	if err != nil {
		return nil, err
	}

	setDefaults(&config)

	return &config, nil

}
