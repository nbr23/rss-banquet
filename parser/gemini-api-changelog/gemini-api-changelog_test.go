package geminiapichangelog

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

func TestGeminiAPIChangelogParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		GeminiAPIChangelog{},
		&parser.Options{
			OptionsList: parser.OptionsList{},
			Parser:      GeminiAPIChangelog{},
		},
		5,
		`.+`,
		`^Gemini API Changelog$`,
	)
}
