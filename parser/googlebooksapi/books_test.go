package googlebooksapi

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

// Amélie Nothomb has been publishing yearly for 30 years. Don't break my tests!
func TestAmelieNothomb(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		Googlebooksapi{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{
					Flag:  "author",
					Type:  "string",
					Value: "Amélie Nothomb",
				},
				&parser.Option{
					Flag:  "language",
					Type:  "string",
					Value: "fr",
				},
			},
		},
		0,
		`^.* - Amélie Nothomb$`,
	)
}
