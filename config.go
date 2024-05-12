package main

import (
	"os"

	"github.com/nbr23/rss-banquet/parser"
	"gopkg.in/yaml.v3"
)

const ENV_PREFIX = "ATOMICBANQUET_"

type FeedConfig struct {
	Module     string          `yaml:"module"`
	OptionsRaw map[string]any  `yaml:"options"`
	Options    *parser.Options `yaml:"-"`
	Name       string          `yaml:"name"`
	FeedType   string          `yaml:"type"`
}

type Config struct {
	Feeds      []*FeedConfig `yaml:"feeds"`
	OutputPath string        `yaml:"output_path"`
	BuildIndex bool          `yaml:"build_index"`
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

	for _, feed := range config.Feeds {
		m := Modules[feed.Module]()
		feed.Options = parser.GetFullOptions(m)
		feed.Options.ParseYaml(feed.OptionsRaw)
	}

	return &config, nil
}
