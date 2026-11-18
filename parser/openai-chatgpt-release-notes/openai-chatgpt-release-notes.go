package openaichatgptreleasenotes

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const url = "https://help.openai.com/en/articles/6825453-chatgpt-release-notes"

var getReleaseNotes = parser.HttpGetUtls

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

func parseBlocks(blocks *goquery.Selection, dated bool, date time.Time) []*feeds.Item {
	var items []*feeds.Item
	var current *section
	var dateStr string
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

	blocks.Each(func(_ int, child *goquery.Selection) {
		switch goquery.NodeName(child) {
		case "h1":
			flush()
			dateStr = ""
			dated = false
			sectionHasTitle = false
			s := cleanText(child.Text())
			if d, err := parseDate(s); err == nil {
				dateStr, date, dated = s, d, true
				current = &section{dateStr: dateStr, date: date}
			}
		case "h2":
			if !dated {
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

var updatedAtRegex = regexp.MustCompile(`\\?"updatedAt\\?",(\d+)`)

func articleUpdatedAt(doc *goquery.Document) (time.Time, bool) {
	m := updatedAtRegex.FindStringSubmatch(doc.Find("script").Text())
	if m == nil {
		return time.Time{}, false
	}
	secs, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(secs, 0).UTC(), true
}

func leadingBlocks(doc *goquery.Document) *goquery.Selection {
	return doc.Find("div.article-content").First().Children().FilterFunction(func(_ int, s *goquery.Selection) bool {
		return !s.Is("div.intercom-interblocks-callout") && s.PrevAllFiltered("div.intercom-interblocks-callout").Length() == 0
	})
}

func newestDate(items []*feeds.Item) (time.Time, bool) {
	var newest time.Time
	for _, item := range items {
		if item.Created.After(newest) {
			newest = item.Created
		}
	}
	return newest, !newest.IsZero()
}

func parseDocument(doc *goquery.Document) []*feeds.Item {
	var items []*feeds.Item
	doc.Find("div.intercom-interblocks-callout").Each(func(_ int, callout *goquery.Selection) {
		items = append(items, parseBlocks(callout.Children(), false, time.Time{})...)
	})

	date, ok := articleUpdatedAt(doc)
	if !ok {
		date, ok = newestDate(items)
	}
	if !ok {
		return items
	}
	// Entries published above the first callout carry no date heading of their own
	return append(parseBlocks(leadingBlocks(doc), true, date), items...)
}

func (OpenAIChatGPTReleaseNotes) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed

	resp, err := getReleaseNotes(url, nil)
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
