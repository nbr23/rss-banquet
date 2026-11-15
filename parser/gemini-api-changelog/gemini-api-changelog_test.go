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

func TestKeyStableAcrossCopyEdits(t *testing.T) {
	pairs := [][2]string{
		{"Released Gemini Robotics ER 1.5 model in preview.", "Released Gemini Robotics-ER 1.5 model in preview."},
		{"Open in Colab &ndash; exports a prompt and the code", "Open in Colab – exports a prompt and the code"},
		{"**Gemini 3.8 Live and Gemini 3.8 Live Extended Thinking generally available (GA)**: Released two new audio-to-audio models.", "**Gemini 3.8 Live and Gemini 3.8 Live Extended Thinking generally available (GA)**: Released our next-generation Live models."},
	}
	for _, p := range pairs {
		a := parser.TextKey(toPlainText([]string{p[0]}))
		b := parser.TextKey(toPlainText([]string{p[1]}))
		if a != b {
			t.Errorf("key mismatch: %q != %q", a, b)
		}
	}
}

func TestDuplicateEntriesGetDistinctGuids(t *testing.T) {
	md := `## February 19, 2025

*   Support for additional regions (Kosovo, Greenland and Faroe Islands).
*   Support for additional regions (Kosovo, Greenland and Faroe Islands).
`
	items := buildItems(parseEntries(md))
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Id == items[1].Id {
		t.Errorf("duplicate guid %s", items[0].Id)
	}
}
