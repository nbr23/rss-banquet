package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func snapshotConfig(t *testing.T) {
	t.Helper()
	saved := make([]string, len(CONFIG_OPTIONS))
	for i := range CONFIG_OPTIONS {
		saved[i] = CONFIG_OPTIONS[i].Value
	}
	t.Cleanup(func() {
		for i := range CONFIG_OPTIONS {
			CONFIG_OPTIONS[i].Value = saved[i]
		}
	})
}

func TestGetConfigOption(t *testing.T) {
	snapshotConfig(t)

	assert.Equal(t, "info", GetConfigOption("LOG_LEVEL"))
	assert.Equal(t, "8080", GetConfigOption("SERVER_PORT"))
	assert.Equal(t, "", GetConfigOption("DOES_NOT_EXIST"))
}

func TestGetConfigOptionIfExists(t *testing.T) {
	snapshotConfig(t)

	value, err := GetConfigOptionIfExists("LOG_LEVEL")
	assert.NoError(t, err)
	assert.Equal(t, "info", value)

	value, err = GetConfigOptionIfExists("DOES_NOT_EXIST")
	assert.Error(t, err)
	assert.Equal(t, "", value)
}

func TestInitConfigFromEnv(t *testing.T) {
	snapshotConfig(t)

	t.Setenv("BANQUET_GLOBAL_LOG_LEVEL", "debug")
	t.Setenv("BANQUET_SERVER_SERVER_PORT", "9090")

	InitConfig()

	assert.Equal(t, "debug", GetConfigOption("LOG_LEVEL"))
	assert.Equal(t, "9090", GetConfigOption("SERVER_PORT"))
}

func TestInitConfigKeepsDefaults(t *testing.T) {
	snapshotConfig(t)

	t.Setenv("BANQUET_GLOBAL_LOG_LEVEL", "")

	InitConfig()

	assert.Equal(t, "info", GetConfigOption("LOG_LEVEL"))
}

func TestReadmeText(t *testing.T) {
	snapshotConfig(t)

	readme := ReadmeText()

	assert.Contains(t, readme, "-  `BANQUET_GLOBAL_LOG_LEVEL`: Log level")
	assert.Contains(t, readme, "(default: info)")
	assert.Contains(t, readme, "-  `BANQUET_GLOBAL_USER_AGENT`:")
	assert.Contains(t, readme, "-  `BANQUET_SERVER_SERVER_PORT`:")
	assert.Contains(t, readme, "(default: 8080)")
}
