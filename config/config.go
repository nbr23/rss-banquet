package config

import (
	"fmt"
	"os"
	"strings"
)

const ENV_PREFIX = "BANQUET"

type ConfigOption struct {
	Name        string
	Value       string
	Scope       string
	Description string
}

var CONFIG_OPTIONS = []ConfigOption{
	{
		Name:        "LOG_LEVEL",
		Value:       "info",
		Scope:       "GLOBAL",
		Description: "Log level (trace, debug, info, warn, error, fatal, panic, disabled)",
	},
	{
		Name:        "USER_AGENT",
		Value:       "",
		Scope:       "GLOBAL",
		Description: "User agent to use for HTTP requests",
	},
	{
		Name:        "SERVER_PORT",
		Value:       "8080",
		Scope:       "SERVER",
		Description: "Port to listen on in server mode",
	},
}

func ReadmeText() string {
	s := "The following environment variables can be used to configure the application:\n\n"
	for _, option := range CONFIG_OPTIONS {
		envVarName := strings.Join([]string{ENV_PREFIX, option.Scope, option.Name}, "_")

		s += fmt.Sprintf("-  `%s`: %s", envVarName, option.Description)
		if option.Value != "" {
			s += fmt.Sprintf(" (default: %s)", option.Value)
		}
		s += "\n"
	}
	return s
}

func InitConfig() {
	for i := range CONFIG_OPTIONS {
		env := os.Getenv(strings.Join([]string{ENV_PREFIX, CONFIG_OPTIONS[i].Scope, CONFIG_OPTIONS[i].Name}, "_"))
		if env != "" {
			CONFIG_OPTIONS[i].Value = env
		}
	}
}

func GetConfigOptionIfExists(name string) (string, error) {
	for _, option := range CONFIG_OPTIONS {
		if option.Name == name {
			return option.Value, nil
		}
	}
	return "", fmt.Errorf("option not found")
}

func GetConfigOption(name string) string {
	for _, option := range CONFIG_OPTIONS {
		if option.Name == name {
			return option.Value
		}
	}
	return ""
}
