package openaiapichangelog

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

func TestOpenAIAPIChangelogParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		OpenAIAPIChangelog{},
		&parser.Options{
			OptionsList: parser.OptionsList{},
			Parser:      OpenAIAPIChangelog{},
		},
		5,
		`.+`,
		`^OpenAI API Changelog$`,
	)
}
