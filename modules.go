package main

import (
	"github.com/nbr23/atomic-banquet/parser"
	"github.com/nbr23/atomic-banquet/parser/books"
	"github.com/nbr23/atomic-banquet/parser/bugcrowd"
	"github.com/nbr23/atomic-banquet/parser/dockerhub"
	garminwearables "github.com/nbr23/atomic-banquet/parser/garmin-wearables"
	"github.com/nbr23/atomic-banquet/parser/garminsdk"
	"github.com/nbr23/atomic-banquet/parser/hackerone"
	"github.com/nbr23/atomic-banquet/parser/hackeronePrograms"
	"github.com/nbr23/atomic-banquet/parser/infocon"
	"github.com/nbr23/atomic-banquet/parser/lego"
	"github.com/nbr23/atomic-banquet/parser/pentesterland"
	"github.com/nbr23/atomic-banquet/parser/psupdates"
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
}

func GetModule(name string) parser.Parser {
	m, ok := Modules[name]
	if ok {
		return m()
	}
	return nil
}
