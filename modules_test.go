package main

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/nbr23/rss-banquet/parser"
)

var moduleRoutes = map[string]string{
	"anthropic-api-changelog":      "anthropic-api-changelog",
	"bugcrowd":                     "bugcrowd",
	"costco":                       "costco",
	"dockerhub":                    "dockerhub",
	"garmin-wearables":             "garminwearables",
	"gemini-api-changelog":         "gemini-api-changelog",
	"github-notifications":         "github-notifications",
	"goodreads":                    "goodreads",
	"hackerone":                    "hackerone",
	"hackeronePrograms":            "hackeroneprograms",
	"imdb":                         "imdb",
	"infocon":                      "infocon",
	"lego":                         "lego",
	"nytimes":                      "nytimes",
	"openai-api-changelog":         "openai-api-changelog",
	"openai-chatgpt-release-notes": "openai-chatgpt-release-notes",
	"pentesterland":                "pentesterland",
	"pocorgtfo":                    "pocorgtfo",
	"psupdates":                    "psupdates",
}

func TestModuleRoutes(t *testing.T) {
	assert.Len(t, Modules, len(moduleRoutes))

	for name := range Modules {
		route, known := moduleRoutes[name]
		if !assert.True(t, known, "module %s is missing from moduleRoutes", name) {
			continue
		}
		p := getModule(name)
		if !assert.NotNil(t, p, name) {
			continue
		}
		assert.Equal(t, route, p.String(), name)
	}
}

func TestGetModuleUnknown(t *testing.T) {
	assert.Nil(t, getModule("does-not-exist"))
}

func TestModuleOptions(t *testing.T) {
	for name := range Modules {
		t.Run(name, func(t *testing.T) {
			options := parser.GetFullOptions(getModule(name))

			seen := make(map[string]bool)
			for _, option := range options.OptionsList {
				assert.False(t, seen[option.Flag], "duplicate flag %s", option.Flag)
				seen[option.Flag] = true

				switch option.Type {
				case "string", "stringSlice", "bool":
				case "int":
					_, err := strconv.Atoi(option.Default)
					assert.NoError(t, err, "int option %s has a non-numeric default", option.Flag)
				default:
					t.Errorf("option %s has unsupported type %q", option.Flag, option.Type)
				}
			}

			assert.True(t, seen["feedFormat"])
			assert.True(t, seen["route"])
		})
	}
}

func TestModulesRegisterUniqueRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := gin.New()

	for name := range Modules {
		p := getModule(name)
		assert.NotPanics(t, func() {
			parser.Route(g, p, parser.GetFullOptions(p))
		}, name)
	}

	paths := make(map[string]bool)
	for _, route := range g.Routes() {
		assert.False(t, paths[route.Path], "duplicate route %s", route.Path)
		paths[route.Path] = true
	}
	assert.Len(t, paths, len(Modules))
}
