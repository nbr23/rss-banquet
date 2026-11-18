package anthropicapichangelog

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

func TestAnthropicAPIChangelogParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		AnthropicAPIChangelog{},
		&parser.Options{
			OptionsList: parser.OptionsList{},
			Parser:      AnthropicAPIChangelog{},
		},
		5,
		`.+`,
		`^Anthropic API Changelog$`,
	)
}
