package main

import (
	"github.com/nbr23/atomic-banquet/parser"
	"github.com/nbr23/atomic-banquet/parser/bugcrowd"
	"github.com/nbr23/atomic-banquet/parser/hackerone"
	"github.com/nbr23/atomic-banquet/parser/lego"
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
	"lego": func() parser.Parser {
		return lego.LegoParser()
	},
}
