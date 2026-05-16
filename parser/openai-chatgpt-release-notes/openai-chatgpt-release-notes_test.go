package openaichatgptreleasenotes

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

func TestOpenAIChatGPTReleaseNotesParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		OpenAIChatGPTReleaseNotes{},
		&parser.Options{
			OptionsList: parser.OptionsList{},
			Parser:      OpenAIChatGPTReleaseNotes{},
		},
		5,
		`.+`,
		`^ChatGPT Release Notes$`,
	)
}
