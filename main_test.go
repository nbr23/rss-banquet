package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nbr23/rss-banquet/config"
)

func snapshotConfig(t *testing.T) {
	t.Helper()
	saved := make([]string, len(config.CONFIG_OPTIONS))
	for i := range config.CONFIG_OPTIONS {
		saved[i] = config.CONFIG_OPTIONS[i].Value
	}
	t.Cleanup(func() {
		for i := range config.CONFIG_OPTIONS {
			config.CONFIG_OPTIONS[i].Value = saved[i]
		}
	})
}

func TestGetRunServerFlagsDefaultPort(t *testing.T) {
	snapshotConfig(t)

	var f runServerFlags
	getRunServerFlags(&f)

	assert.Equal(t, "8080", f.serverPort)
}

func TestGetRunServerFlagsPortFromConfig(t *testing.T) {
	snapshotConfig(t)

	t.Setenv("BANQUET_SERVER_SERVER_PORT", "9090")
	config.InitConfig()

	var f runServerFlags
	getRunServerFlags(&f)

	assert.Equal(t, "9090", f.serverPort)
}

func TestGetRunServerFlagsPortFlagOverridesConfig(t *testing.T) {
	snapshotConfig(t)

	t.Setenv("BANQUET_SERVER_SERVER_PORT", "9090")
	config.InitConfig()

	var f runServerFlags
	flags := getRunServerFlags(&f)
	assert.NoError(t, flags.Parse([]string{"-p", "1234", "-h"}))

	assert.Equal(t, "1234", f.serverPort)
	assert.True(t, f.showHelp)
}
