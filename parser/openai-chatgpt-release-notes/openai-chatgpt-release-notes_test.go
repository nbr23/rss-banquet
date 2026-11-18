package openaichatgptreleasenotes

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

func TestOpenAIChatGPTReleaseNotesParse(t *testing.T) {
	previousGet := getReleaseNotes
	getReleaseNotes = func(gotURL string, options map[string]any) (*http.Response, error) {
		if gotURL != url {
			t.Fatalf("got URL %q, want %q", gotURL, url)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(fixture)),
		}, nil
	}
	t.Cleanup(func() { getReleaseNotes = previousGet })

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

const fixture = `<html><body>
<div class="intercom-interblocks-callout">
<h1>December 17th, 2025</h1>
<h2>Tasks are now in Pulse</h2>
<p>Tasks body.</p>
<h1>December 16, 2025</h1>
<h2>ChatGPT Images on Web and Mobile</h2>
<p>Images body.</p>
<h2>Second item</h2>
<p>Second body.</p>
<h1>December 11, 2025</h1>
<p>Today, we're releasing <b>GPT-5.2</b>, the next upgrade. More text.</p>
<h1>October 23, 2025</h1>
<h3>Shared projects</h3>
<p>Starting today, project sharing is available.</p>
<h1>February 14, 2025</h1>
<ul><li><p>Set ChatGPT as default search engine in Safari.</p></li><li><p>Confirmation dialog for logging out.</p></li></ul>
<h1>February 21, 2025</h1>
<p>1&#x2060;. iOS widget (see below):</p>
</div>
<div class="intercom-interblocks-callout">
<h1>Feb 9 2023</h1>
<p>Intro paragraph dropped.</p>
<h2>Introducing ChatGPT Plus</h2>
<p>Plus body.</p>
</div>
<div class="intercom-interblocks-callout">
<h1>Apr 1, 2024</h1>
<h2>Start using ChatGPT instantly</h2>
<p>Instant body.</p>
</div>
</body></html>`

func TestParseDocument(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fixture))
	if err != nil {
		t.Fatal(err)
	}
	items := parseDocument(doc)

	want := []struct {
		dateStr string
		date    string
		title   string
	}{
		{"December 17th, 2025", "2025-12-17", "Tasks are now in Pulse"},
		{"December 16, 2025", "2025-12-16", "ChatGPT Images on Web and Mobile"},
		{"December 16, 2025", "2025-12-16", "Second item"},
		{"December 11, 2025", "2025-12-11", "Today, we're releasing GPT-5.2, the next upgrade"},
		{"October 23, 2025", "2025-10-23", "Shared projects"},
		{"February 14, 2025", "2025-02-14", "Set ChatGPT as default search engine in Safari."},
		{"February 21, 2025", "2025-02-21", "ChatGPT update: February 21, 2025"},
		{"Feb 9 2023", "2023-02-09", "Introducing ChatGPT Plus"},
		{"Apr 1, 2024", "2024-04-01", "Start using ChatGPT instantly"},
	}
	if len(items) != len(want) {
		t.Fatalf("got %d items, want %d", len(items), len(want))
	}
	for i, w := range want {
		item := items[i]
		if item.Title != w.title {
			t.Errorf("item %d: title %q, want %q", i, item.Title, w.title)
		}
		if got := item.Created.Format("2006-01-02"); got != w.date {
			t.Errorf("item %d: date %s, want %s", i, got, w.date)
		}
		if id := parser.GetGuid([]string{w.dateStr, w.title}); item.Id != id {
			t.Errorf("item %d: unexpected guid", i)
		}
	}
	if strings.Contains(items[7].Description, "Intro paragraph") {
		t.Errorf("intro before first h2 leaked into item: %q", items[7].Description)
	}
}

const leadingFixture = `<html><body>
<div class="article-content">
<h2>Two new models in Work</h2>
<p>GPT-6.2 Astra follows your template more closely.</p>
<div class="intercom-interblocks-callout">
<h1>November 16, 2026</h1>
<h2>Clinical search</h2>
<p>Clinical body.</p>
</div>
<h2>Trailing heading</h2>
<p>Not a leading entry.</p>
</div>
%s
</body></html>`

func TestParseDocumentLeadingEntry(t *testing.T) {
	cases := []struct {
		name   string
		script string
		date   string
	}{
		{"updatedAt", `<script>self.__next_f.push([1,"\"updatedAt\",1794937479.286,\"fetchedAt\""])</script>`, "2026-11-17"},
		{"fallback", "", "2026-11-16"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(fmt.Sprintf(leadingFixture, c.script)))
			if err != nil {
				t.Fatal(err)
			}
			items := parseDocument(doc)
			if len(items) != 2 {
				t.Fatalf("got %d items, want 2", len(items))
			}
			lead := items[0]
			if lead.Title != "Two new models in Work" {
				t.Errorf("title %q", lead.Title)
			}
			if got := lead.Created.Format("2006-01-02"); got != c.date {
				t.Errorf("date %s, want %s", got, c.date)
			}
			if !strings.Contains(lead.Description, "GPT-6.2 Astra") {
				t.Errorf("description %q", lead.Description)
			}
			if items[1].Title != "Clinical search" {
				t.Errorf("second item title %q", items[1].Title)
			}
		})
	}
}
