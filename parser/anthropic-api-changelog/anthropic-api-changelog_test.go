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

func TestIdKeyStableAcrossCopyEdits(t *testing.T) {
	pairs := [][2]string{
		{"We've launched the Files API in public beta, enabling you to upload files", "We’ve launched the Files API in beta, enabling you to upload files"},
		{"Claude Opus 4.7 supports high-resolution image input, raising the maximum image resolution from 1568 to 2576 pixels", "Claude Opus 4.7 supports high-resolution image input, raising the maximum image resolution from 1,568 to 2,576 pixels"},
		{"Some customers might encounter 429 (rate_limit_error) errors following a sharp increase in API usage due to acceleration limits", "Some customers might encounter 429 (rate_limit_error) errors following a sharp increase in API usage because of acceleration limits"},
	}
	for _, p := range pairs {
		if a, b := idKey(p[0]), idKey(p[1]); a != b {
			t.Errorf("idKey mismatch: %q != %q", a, b)
		}
	}
}
