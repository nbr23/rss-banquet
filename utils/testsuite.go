package testsuite

import (
	"regexp"
	"testing"

	"github.com/nbr23/rss-banquet/parser"
)

func TestParseSuccess(t *testing.T,
	p parser.Parser,
	parserOptions *parser.Options,
	minItem int,
	itemTitleRegex string,
) {
	parsed, err := p.Parse(parserOptions)
	if err != nil {
		t.Errorf("Unable to parse: %s", err)
		return
	}

	if parsed == nil {
		t.Errorf("Unable to parse: nil feed returned")
		return
	}

	if parsed.Items == nil || len(parsed.Items) < minItem {
		t.Errorf("Unable to parse: not enough items in feed")
		return
	}

	if itemTitleRegex != "" {
		r := regexp.MustCompile(itemTitleRegex)
		if !r.MatchString(parsed.Items[0].Title) {
			t.Errorf("Unable to parse, title doesn't match expected format, got '%s'", parsed.Items[0].Title)
		}
	}
}

func TestParseFailure(t *testing.T,
	p parser.Parser,
	parserOptions *parser.Options,
) {
	_, err := p.Parse(parserOptions)

	if err == nil || err.Error() != "unable to fetch the update page, status code: 404" {
		t.Errorf("Failed to fail on bad options: %s", err)
		return
	}
}
