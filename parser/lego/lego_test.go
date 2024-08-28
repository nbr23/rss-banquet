package lego

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

func TestLegoParseNew(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		Lego{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{
					Flag:  "category",
					Type:  "string",
					Value: "new",
				},
			},
			Parser: Lego{},
		},
		1,
		`^\[\w+\] [0-9]+ - .* - (Available now|Pre-order this item today,|Will ship by|Backorder) .*$`,
	)
}

func TestLegoParseComingSoon(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		Lego{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{
					Flag:  "category",
					Type:  "string",
					Value: "coming-soon",
				},
			},
			Parser: Lego{},
		},
		1,
		`^\[COMING SOON\] [0-9]+ - .* - Coming Soon .*$`,
	)
}
