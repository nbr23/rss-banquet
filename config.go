package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const ENV_PREFIX = "ATOMICBANQUET_"

type FeedConfig struct {
	Module  string         `yaml:"module"`
	Options map[string]any `yaml:"options"`
	Name    string         `yaml:"name"`
}

type Config struct {
	Feeds      []FeedConfig `yaml:"feeds"`
	OutputPath string       `yaml:"output_path"`
	BuildIndex bool         `yaml:"build_index"`
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

	extendConfigFromEnv(&config)

	return &config, nil

}

func extendConfigFromEnv(config *Config) {
	feedOptPrefix := fmt.Sprintf("%sFEED_OPT_", ENV_PREFIX)
	for _, env := range os.Environ() {
		keyval := strings.SplitN(env, "=", 2)
		k, v := keyval[0], keyval[1]

		if strings.HasPrefix(k, feedOptPrefix) {
			p := strings.Split(strings.TrimPrefix(k, feedOptPrefix), "_")
			feedKey := strings.ToLower(p[0])
			optionKey := p[1]
			for _, feed := range config.Feeds {
				if feed.Module == feedKey {
					feed.Options[optionKey] = v
				}
			}
		}
	}
}
