package main

import (
	"github.com/nbr23/atomic-banquet/parser"
	"github.com/nbr23/atomic-banquet/parser/bugcrowd"
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
}
