package openaichatgptreleasenotes

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const url = "https://help.openai.com/en/articles/6825453-chatgpt-release-notes"

type OpenAIChatGPTReleaseNotes struct{}

func (OpenAIChatGPTReleaseNotes) String() string {
	return "openai-chatgpt-release-notes"
}

func (OpenAIChatGPTReleaseNotes) GetOptions() parser.Options {
	return parser.Options{}
}

var (
	ordinalRegex = regexp.MustCompile(`(\d+)(st|nd|rd|th)`)
	dateLayouts  = []string{"January 2, 2006", "January 2 2006", "Jan 2, 2006", "Jan 2 2006"}
)

func normalizeDate(s string) string {
	return ordinalRegex.ReplaceAllString(s, "$1")
}

func parseDate(s string) (time.Time, error) {
	s = normalizeDate(s)
	var err error
	for _, layout := range dateLayouts {
		var d time.Time
		if d, err = time.Parse(layout, s); err == nil {
			return d, nil
		}
	}
	return time.Time{}, err
}

func cleanText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

type section struct {
	dateStr string
	date    time.Time
	title   string
	lead    string
	html    strings.Builder
	text    strings.Builder
}

func (s *section) append(child *goquery.Selection) {
	if html, err := goquery.OuterHtml(child); err == nil {
		s.html.WriteString(html)
	}
	text := strings.TrimSpace(child.Text())
	if text == "" {
		return
	}
	if s.lead == "" {
		s.lead = leadText(child)
	}
	if s.text.Len() > 0 {
		s.text.WriteString("\n\n")
	}
	s.text.WriteString(text)
}

func leadText(child *goquery.Selection) string {
	blocks := child.Filter("h3, h4, p").AddSelection(child.Find("h3, h4, p"))
	for i := range blocks.Nodes {
		if text := cleanText(blocks.Eq(i).Text()); text != "" {
			return text
		}
	}
	return cleanText(child.Text())
}

func headlessTitle(lead, dateStr string) string {
	title := parser.ExtractTitle(lead)
	if len(strings.Fields(parser.TextKey(title))) < 2 {
		return "ChatGPT update: " + dateStr
	}
	return title
}

func (s *section) item() *feeds.Item {
	text := strings.TrimSpace(s.text.String())
	content := strings.TrimSpace(s.html.String())
	if content == "" {
		content = text
	}
	title := s.title
	if title == "" {
		if text == "" {
			return nil
		}
		title = headlessTitle(s.lead, s.dateStr)
	}
	return &feeds.Item{
		Title:       title,
		Content:     content,
		Description: text,
		Link:        &feeds.Link{Href: url},
		Created:     s.date,
		Id:          parser.GetGuid([]string{s.dateStr, title}),
	}
}

func parseCallout(callout *goquery.Selection) []*feeds.Item {
	var items []*feeds.Item
	var current *section
	var dateStr string
	var date time.Time
	sectionHasTitle := false

	flush := func() {
		if current == nil {
			return
		}
		if item := current.item(); item != nil {
			items = append(items, item)
		}
		current = nil
	}

	callout.Children().Each(func(_ int, child *goquery.Selection) {
		switch goquery.NodeName(child) {
		case "h1":
			flush()
			dateStr = ""
			sectionHasTitle = false
			s := cleanText(child.Text())
			if d, err := parseDate(s); err == nil {
				dateStr, date = s, d
				current = &section{dateStr: dateStr, date: date}
			}
		case "h2":
			if dateStr == "" {
				return
			}
			if sectionHasTitle {
				flush()
			}
			sectionHasTitle = true
			current = &section{dateStr: dateStr, date: date, title: cleanText(child.Text())}
		default:
			if current != nil {
				current.append(child)
			}
		}
	})
	flush()

	return items
}

func parseDocument(doc *goquery.Document) []*feeds.Item {
	var items []*feeds.Item
	doc.Find("div.intercom-interblocks-callout").Each(func(_ int, callout *goquery.Selection) {
		items = append(items, parseCallout(callout)...)
	})
	return items
}

func (OpenAIChatGPTReleaseNotes) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed

	resp, err := parser.HttpGetUtls(url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, parser.NewInternalError(fmt.Sprintf("unable to fetch the release notes page, status code: %d", resp.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	feed.Items = parseDocument(doc)
	feed.Title = "ChatGPT Release Notes"
	feed.Description = "Release notes for ChatGPT"
	feed.Author = &feeds.Author{Name: "OpenAI"}
	feed.Link = &feeds.Link{Href: url}

	parser.SortFeedEntries(&feed)

	return &feed, nil
}

func OpenAIChatGPTReleaseNotesParser() parser.Parser {
	return OpenAIChatGPTReleaseNotes{}
}
