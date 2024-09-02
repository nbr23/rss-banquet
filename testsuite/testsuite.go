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
	feedTitleRegex string,
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

	if len(parsed.Items) < minItem && minItem != 0 {
		t.Errorf("Unable to parse: not enough items in feed")
		return
	}

	if itemTitleRegex != "" && len(parsed.Items) > 0 {
		r := regexp.MustCompile(itemTitleRegex)
		if !r.MatchString(parsed.Items[0].Title) {
			t.Errorf("Unable to parse, title doesn't match expected format, got '%s'", parsed.Items[0].Title)
		}
	}

	if feedTitleRegex != "" {
		r := regexp.MustCompile(feedTitleRegex)
		if !r.MatchString(parsed.Title) {
			t.Errorf("Unable to parse, feed title doesn't match expected format, got '%s'", parsed.Title)
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
