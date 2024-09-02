package config

import (
	"fmt"
	"os"
	"strings"
)

const ENV_PREFIX = "BANQUET"

type ConfigOption struct {
	Name  string
	Value string
	Scope string
}

var CONFIG_OPTIONS = []ConfigOption{
	{
		Name:  "LOG_LEVEL",
		Value: "info",
		Scope: "GLOBAL",
	},
	{
		Name:  "USER_AGENT",
		Value: "",
		Scope: "GLOBAL",
	},
}

func InitConfig() {
	for i := range CONFIG_OPTIONS {
		env := os.Getenv(strings.Join([]string{ENV_PREFIX, CONFIG_OPTIONS[i].Scope, CONFIG_OPTIONS[i].Name}, "_"))
		if env != "" {
			CONFIG_OPTIONS[i].Value = env
		}
	}
}

func GetConfigOption(name string) (string, error) {
	for _, option := range CONFIG_OPTIONS {
		if option.Name == name {
			return option.Value, nil
		}
	}
	return "", fmt.Errorf("option not found")
}
