package psupdates

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/utils"
)

func TestPS5UpdatesParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		PSUpdates{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "hardware", Type: "string", Value: "ps5"},
				&parser.Option{Flag: "local", Type: "string", Value: "en-us"},
			},
			Parser: PSUpdates{},
		},
		1,
		`^PS5 Update: [^\s]+.*$`,
	)
}

func TestPS4UpdatesParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		PSUpdates{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "hardware", Type: "string", Value: "ps4"},
				&parser.Option{Flag: "local", Type: "string", Value: "en-us"},
			},
			Parser: PSUpdates{},
		},
		1,
		`^PS4 Update: [^\s]+.*$`,
	)
}

func TestPSUpdatesBadOptions(t *testing.T) {
	testsuite.TestParseFailure(
		t,
		PSUpdates{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "hardware", Type: "string", Default: "ps5", Value: "ps1"},
				&parser.Option{Flag: "local", Type: "string", Default: "en-us", Value: "en-us"},
			},
			Parser: PSUpdates{},
		},
	)
}
