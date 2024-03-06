package books

import (
	"testing"

	"github.com/nbr23/atomic-banquet/parser"
	testsuite "github.com/nbr23/atomic-banquet/utils"
)

// Amélie Nothomb has been publishing yearly for 30 years. Don't break my tests!
func TestAmelieNothomb(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		Books{},
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
		1,
		`^.* - Amélie Nothomb$`,
	)
}
