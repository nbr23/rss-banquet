package main

import (
	"fmt"
	"sort"

	"github.com/nbr23/rss-banquet/parser"
	anthropicapichangelog "github.com/nbr23/rss-banquet/parser/anthropic-api-changelog"
	"github.com/nbr23/rss-banquet/parser/bugcrowd"
	"github.com/nbr23/rss-banquet/parser/costco"
	"github.com/nbr23/rss-banquet/parser/dockerhub"
	garminwearables "github.com/nbr23/rss-banquet/parser/garmin-wearables"
	geminiapichangelog "github.com/nbr23/rss-banquet/parser/gemini-api-changelog"
	githubnotifications "github.com/nbr23/rss-banquet/parser/github-notifications"
	"github.com/nbr23/rss-banquet/parser/goodreads"
	"github.com/nbr23/rss-banquet/parser/hackerone"
	"github.com/nbr23/rss-banquet/parser/hackeronePrograms"
	"github.com/nbr23/rss-banquet/parser/imdb"
	"github.com/nbr23/rss-banquet/parser/infocon"
	"github.com/nbr23/rss-banquet/parser/lego"
	"github.com/nbr23/rss-banquet/parser/nytimes"
	openaiapichangelog "github.com/nbr23/rss-banquet/parser/openai-api-changelog"
	openaichatgptreleasenotes "github.com/nbr23/rss-banquet/parser/openai-chatgpt-release-notes"
	"github.com/nbr23/rss-banquet/parser/pentesterland"
	"github.com/nbr23/rss-banquet/parser/pocorgtfo"
	"github.com/nbr23/rss-banquet/parser/psupdates"
)

var Modules = map[string]func() parser.Parser{
	"psupdates": func() parser.Parser {
		return psupdates.PSUpdatesParser()
	},
	"bugcrowd": func() parser.Parser {
		return bugcrowd.BugcrowdParser()
	},
	"hackerone": func() parser.Parser {
		return hackerone.HackeroneParser()
	},
	"hackeronePrograms": func() parser.Parser {
		return hackeronePrograms.HackeroneProgramsParser()
	},
	"lego": func() parser.Parser {
		return lego.LegoParser()
	},
	"infocon": func() parser.Parser {
		return infocon.InfoConParser()
	},
	"imdb": func() parser.Parser {
		return imdb.IMDBParser()
	},
	"pentesterland": func() parser.Parser {
		return pentesterland.PentesterLandParser()
	},
	"garmin-wearables": func() parser.Parser {
		return garminwearables.GarminWearablesParser()
	},
	"dockerhub": func() parser.Parser {
		return dockerhub.DockerHubParser()
	},
	"pocorgtfo": func() parser.Parser {
		return pocorgtfo.PoCOrGTFOParser()
	},
	"goodreads": func() parser.Parser {
		return goodreads.GoodReadsParser()
	},
	"costco": func() parser.Parser {
		return costco.CostcoParser()
	},
	"nytimes": func() parser.Parser {
		return nytimes.NYTimesParser()
	},
	"github-notifications": func() parser.Parser {
		return githubnotifications.GitHubNotificationsParser()
	},
	"gemini-api-changelog": func() parser.Parser {
		return geminiapichangelog.GeminiAPIChangelogParser()
	},
	"anthropic-api-changelog": func() parser.Parser {
		return anthropicapichangelog.AnthropicAPIChangelogParser()
	},
	"openai-api-changelog": func() parser.Parser {
		return openaiapichangelog.OpenAIAPIChangelogParser()
	},
	"openai-chatgpt-release-notes": func() parser.Parser {
		return openaichatgptreleasenotes.OpenAIChatGPTReleaseNotesParser()
	},
}

func getModule(name string) parser.Parser {
	m, ok := Modules[name]
	if ok {
		return m()
	}
	return nil
}

func printModulesHelp() {
	sortedModules := make([]string, 0, len(Modules))
	for key := range Modules {
		sortedModules = append(sortedModules, key)
	}
	sort.Strings(sortedModules)

	for _, module := range sortedModules {
		fmt.Printf("  - %s\n%s\n", module, parser.GetFullOptions(Modules[module]()).GetHelp())
	}
}
