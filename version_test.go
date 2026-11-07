package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nbr23/rss-banquet/parser"
)

func TestResolveBuildInfo(t *testing.T) {
	savedVersion, savedCommit, savedTime := version, commit, commitTime
	t.Cleanup(func() { version, commit, commitTime = savedVersion, savedCommit, savedTime })
	version, commit, commitTime = "", "", ""

	info := &debug.BuildInfo{Settings: []debug.BuildSetting{
		{Key: "vcs.revision", Value: "abc123"},
		{Key: "vcs.time", Value: "2026-11-06T12:00:00Z"},
		{Key: "vcs.modified", Value: "true"},
	}}
	build := resolveBuildInfo(info)
	assert.Equal(t, "v0.0.0-dev+gabc123", build.version)
	assert.Equal(t, "abc123", build.commit)
	assert.Equal(t, time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC), build.committedAt)
	assert.Contains(t, build.String(), "(modified)")

	version, commit, commitTime = "v1.2.0", "def456", "2026-11-07T10:00:00Z"
	build = resolveBuildInfo(info)
	assert.Equal(t, "v1.2.0", build.version)
	assert.Equal(t, "def456", build.commit)
	assert.Equal(t, time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC), build.committedAt)

	version, commit, commitTime = "", "", ""
	build = resolveBuildInfo(nil)
	assert.Equal(t, "v0.0.0-dev", build.version)
	assert.Equal(t, "unknown", build.commit)
	assert.Equal(t, time.Unix(0, 0).UTC(), build.committedAt)
	assert.False(t, build.modified)
	assert.NotContains(t, build.String(), "(modified)")

	info.Main.Version = "v1.3.0"
	assert.Equal(t, "v1.3.0", resolveBuildInfo(info).version)
	info.Main.Version = "(devel)"
	assert.Equal(t, "v0.0.0-dev+gabc123", resolveBuildInfo(info).version)
}

func TestBuildVersionScript(t *testing.T) {
	script, err := filepath.Abs("scripts/version.sh")
	require.NoError(t, err)
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "git"), []byte(`#!/usr/bin/env bash
set -euo pipefail
case "$1" in
    tag) printf '%s\n' "$BANQUET_TEST_TAGS" ;;
    describe)
        for candidate in "$@"; do
            if [[ -n "$BANQUET_TEST_TAG" && "$candidate" == "$BANQUET_TEST_TAG" ]]; then
                printf '%s\n' "$candidate"
                exit 0
            fi
        done
        exit 1
        ;;
    rev-list) printf '%s\n' "$BANQUET_TEST_COUNT" ;;
    rev-parse) printf '%s\n' 'abcdef123456' ;;
    *) exit 1 ;;
esac
`), 0755))

	for _, tc := range []struct {
		name  string
		tags  string
		tag   string
		count string
		want  string
	}{
		{name: "no tags", count: "12", want: "v0.0.0-dev.12+gabcdef123456"},
		{name: "release", tags: "v1.2.3", tag: "v1.2.3", count: "0", want: "v1.2.3"},
		{name: "after release", tags: "v1.2.3", tag: "v1.2.3", count: "4", want: "v1.2.3-dev.4+gabcdef123456"},
		{name: "prerelease", tags: "v1.3.0-rc.1", tag: "v1.3.0-rc.1", count: "0", want: "v1.3.0-rc.1"},
		{name: "after prerelease", tags: "v1.3.0-rc.1", tag: "v1.3.0-rc.1", count: "2", want: "v1.3.0-rc.1.dev.2+gabcdef123456"},
		{name: "build metadata", tags: "v1.2.3+build.1", tag: "v1.2.3+build.1", count: "2", want: "v1.2.3-dev.2+gabcdef123456"},
		{name: "unreachable tag", tags: "v9.0.0", count: "12", want: "v0.0.0-dev.12+gabcdef123456"},
		{name: "unrelated tag", tags: "backup", tag: "backup", count: "12", want: "v0.0.0-dev.12+gabcdef123456"},
		{name: "invalid semver", tags: "v01.2.3\nv1.2\nv1.2.3-01", tag: "v01.2.3", count: "12", want: "v0.0.0-dev.12+gabcdef123456"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("bash", script)
			cmd.Env = append(os.Environ(),
				"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"BANQUET_TEST_TAGS="+tc.tags,
				"BANQUET_TEST_TAG="+tc.tag,
				"BANQUET_TEST_COUNT="+tc.count,
			)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", output)
			assert.Equal(t, tc.want+"\n", string(output))
		})
	}
}

func TestDeploymentFeedIdentity(t *testing.T) {
	build := buildInfo{
		version:     "v0.1.0",
		commit:      "abc123",
		committedAt: time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
	}
	p := deploymentFeed{build: build}
	feed, err := p.Parse(nil)
	require.NoError(t, err)
	require.Len(t, feed.Items, 1)
	item := feed.Items[0]
	assert.Equal(t, "rss-banquet v0.1.0 (abc123)", item.Title)
	assert.Equal(t, "https://github.com/nbr23/rss-banquet/commit/abc123", item.Link.Href)
	assert.Equal(t, build.committedAt, item.Created)

	again, err := (deploymentFeed{build: build}).Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, feed, again)

	for _, change := range []func(*buildInfo){
		func(b *buildInfo) { b.commit = "def456" },
		func(b *buildInfo) { b.version = "v0.2.0" },
		func(b *buildInfo) { b.modified = true },
	} {
		next := build
		change(&next)
		updated, err := (deploymentFeed{build: next}).Parse(nil)
		require.NoError(t, err)
		assert.NotEqual(t, item.Id, updated.Items[0].Id)
	}

	build.commit = "unknown"
	feed, err = (deploymentFeed{build: build}).Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, repositoryURL, feed.Items[0].Link.Href)
}

func TestDeploymentFeedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := gin.New()
	p := deploymentFeed{build: buildInfo{
		version:     "v0.1.0",
		commit:      "abc123",
		committedAt: time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
	}}
	parser.Route(g, p, parser.GetFullOptions(p))

	for _, format := range []string{"", "rss", "atom", "json", "text"} {
		t.Run(format, func(t *testing.T) {
			url := "/feed/rss-banquet?feedFormat=" + format
			w := httptest.NewRecorder()
			g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
			require.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "rss-banquet v0.1.0 (abc123)")
			assert.Contains(t, w.Body.String(), repositoryURL+"/commit/abc123")
			switch format {
			case "json":
				assert.True(t, json.Valid(w.Body.Bytes()))
			case "text":
			default:
				var document any
				assert.NoError(t, xml.Unmarshal(w.Body.Bytes(), &document))
			}
			again := httptest.NewRecorder()
			g.ServeHTTP(again, httptest.NewRequest(http.MethodGet, url, nil))
			assert.Equal(t, w.Body.String(), again.Body.String())
		})
	}
}
