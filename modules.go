package main

import (
	"fmt"
	"sort"

	"github.com/nbr23/rss-banquet/parser"
	"github.com/nbr23/rss-banquet/parser/books"
	"github.com/nbr23/rss-banquet/parser/bugcrowd"
	"github.com/nbr23/rss-banquet/parser/dockerhub"
	garminwearables "github.com/nbr23/rss-banquet/parser/garmin-wearables"
	"github.com/nbr23/rss-banquet/parser/garminsdk"
	"github.com/nbr23/rss-banquet/parser/hackerone"
	"github.com/nbr23/rss-banquet/parser/hackeronePrograms"
	"github.com/nbr23/rss-banquet/parser/infocon"
	"github.com/nbr23/rss-banquet/parser/lego"
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
	"pentesterland": func() parser.Parser {
		return pentesterland.PentesterLandParser()
	},
	"garmin-sdk": func() parser.Parser {
		return garminsdk.GarminSDKParser()
	},
	"garmin-wearables": func() parser.Parser {
		return garminwearables.GarminWearablesParser()
	},
	"dockerhub": func() parser.Parser {
		return dockerhub.DockerHubParser()
	},
	"books": func() parser.Parser {
		return books.BooksParser()
	},
	"pocorgtfo": func() parser.Parser {
		return pocorgtfo.PoCOrGTFOParser()
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
