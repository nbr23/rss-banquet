package main

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

var version string

var commit string
var commitTime string

const repositoryURL = "https://github.com/nbr23/rss-banquet"

type buildInfo struct {
	version     string
	commit      string
	committedAt time.Time
	modified    bool
}

func getBuildInfo() buildInfo {
	info, _ := debug.ReadBuildInfo()
	return resolveBuildInfo(info)
}

func resolveBuildInfo(info *debug.BuildInfo) buildInfo {
	build := buildInfo{version: strings.TrimSpace(version), commit: commit}
	timestamp := commitTime
	if info != nil {
		if build.version == "" && info.Main.Version != "(devel)" {
			build.version = info.Main.Version
		}
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if build.commit == "" {
					build.commit = setting.Value
				}
			case "vcs.time":
				if timestamp == "" {
					timestamp = setting.Value
				}
			case "vcs.modified":
				build.modified = setting.Value == "true"
			}
		}
	}
	build.committedAt, _ = time.Parse(time.RFC3339, timestamp)
	if build.committedAt.IsZero() {
		build.committedAt = time.Unix(0, 0).UTC()
	}
	if build.commit == "" {
		build.commit = "unknown"
	}
	if build.version == "" {
		build.version = "v0.0.0-dev"
		if build.commit != "unknown" {
			revision := build.commit
			if len(revision) > 12 {
				revision = revision[:12]
			}
			build.version += "+g" + revision
		}
	}
	return build
}

func (b buildInfo) String() string {
	label := fmt.Sprintf("rss-banquet %s (%s)", b.version, b.commit)
	if b.modified {
		label += " (modified)"
	}
	return label
}

type deploymentFeed struct {
	build buildInfo
}

func (deploymentFeed) String() string {
	return "rss-banquet"
}

func (deploymentFeed) GetOptions() parser.Options {
	return parser.Options{}
}

func (p deploymentFeed) Parse(*parser.Options) (*feeds.Feed, error) {
	b := p.build
	link := repositoryURL
	if b.commit != "unknown" {
		link += "/commit/" + b.commit
	}
	return &feeds.Feed{
		Id:          repositoryURL,
		Title:       "RSS Banquet deployment",
		Description: "The version and commit running on this RSS Banquet instance",
		Author:      &feeds.Author{Name: "RSS Banquet"},
		Link:        &feeds.Link{Href: repositoryURL},
		Created:     b.committedAt,
		Updated:     b.committedAt,
		Items: []*feeds.Item{{
			Id:          "urn:sha256:" + parser.GetGuid([]string{b.version, b.commit, fmt.Sprint(b.modified)}),
			IsPermaLink: "false",
			Title:       b.String(),
			Description: b.String(),
			Link:        &feeds.Link{Href: link},
			Created:     b.committedAt,
			Updated:     b.committedAt,
		}},
	}, nil
}
