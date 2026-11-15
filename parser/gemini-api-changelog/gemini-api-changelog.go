package geminiapichangelog

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const (
	baseUrl     = "https://aistudio.google.com"
	url         = baseUrl + "/docs/changelog"
	markdownUrl = url + ".md"
)

var (
	dateHeadingRe = regexp.MustCompile(`^## ([^{]+?)\s*(\{%.*%\})?\s*$`)
	topBulletRe   = regexp.MustCompile(`^[*-]\s+`)
	subBulletRe   = regexp.MustCompile(`^\s+[*-]\s+`)
	linkRe        = regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]+)\)`)
	codeRe        = regexp.MustCompile("`([^`]+)`")
	boldRe        = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	inlineBullet  = regexp.MustCompile(`\s\*\s{2,}`)
	dateLayouts   = []string{"January 2, 2006", "January 2 2006"}
)

type GeminiAPIChangelog struct{}

func (GeminiAPIChangelog) String() string {
	return "gemini-api-changelog"
}

func (GeminiAPIChangelog) GetOptions() parser.Options {
	return parser.Options{}
}

func toPlainText(lines []string) string {
	stripped := make([]string, len(lines))
	for i, line := range lines {
		stripped[i] = subBulletRe.ReplaceAllString(line, "")
	}
	text := inlineBullet.ReplaceAllString(html.UnescapeString(strings.Join(stripped, " ")), " ")
	text = linkRe.ReplaceAllString(text, "$1")
	text = codeRe.ReplaceAllString(text, "$1")
	text = boldRe.ReplaceAllString(text, "$1")
	return strings.Join(strings.Fields(text), " ")
}

func absoluteLink(href string) string {
	if strings.HasPrefix(href, "/") {
		return baseUrl + href
	}
	return href
}

func parseDate(s string) (time.Time, error) {
	var err error
	for _, layout := range dateLayouts {
		var d time.Time
		if d, err = time.Parse(layout, s); err == nil {
			return d, nil
		}
	}
	return time.Time{}, err
}

func renderInline(md string) string {
	out := html.EscapeString(strings.Join(strings.Fields(md), " "))
	out = linkRe.ReplaceAllStringFunc(out, func(m string) string {
		parts := linkRe.FindStringSubmatch(m)
		return fmt.Sprintf(`<a href="%s">%s</a>`, absoluteLink(html.UnescapeString(parts[2])), parts[1])
	})
	out = codeRe.ReplaceAllString(out, "<code>$1</code>")
	return boldRe.ReplaceAllString(out, "<strong>$1</strong>")
}

func renderItem(lines []string) string {
	var blocks []string
	var current []string
	var subItems []string

	flushParagraph := func() {
		if len(current) > 0 {
			blocks = append(blocks, "<p>"+renderInline(strings.Join(current, " "))+"</p>")
			current = nil
		}
	}
	flushList := func() {
		if len(subItems) > 0 {
			blocks = append(blocks, "<ul><li>"+strings.Join(subItems, "</li><li>")+"</li></ul>")
			subItems = nil
		}
	}

	inSubItem := false
	for _, line := range lines {
		switch {
		case strings.TrimSpace(line) == "":
			if inSubItem {
				subItems[len(subItems)-1] = renderInline(strings.Join(current, " "))
				current = nil
				inSubItem = false
			}
			flushParagraph()
		case subBulletRe.MatchString(line):
			if inSubItem {
				subItems[len(subItems)-1] = renderInline(strings.Join(current, " "))
			} else {
				flushParagraph()
			}
			current = []string{subBulletRe.ReplaceAllString(line, "")}
			subItems = append(subItems, "")
			inSubItem = true
		default:
			if !inSubItem {
				flushList()
			}
			current = append(current, strings.TrimSpace(line))
		}
	}
	if inSubItem {
		subItems[len(subItems)-1] = renderInline(strings.Join(current, " "))
		current = nil
	}
	flushParagraph()
	flushList()

	return strings.Join(blocks, "")
}

type entry struct {
	date  time.Time
	lines []string
}

func parseEntries(md string) []entry {
	var entries []entry
	var current *entry
	var date time.Time
	var dateStr string
	inList := false

	flush := func() {
		if current != nil {
			entries = append(entries, *current)
			current = nil
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(md))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			flush()
			inList = false
			dateStr = ""
			if m := dateHeadingRe.FindStringSubmatch(line); m != nil {
				if d, err := parseDate(m[1]); err == nil {
					date, dateStr = d, m[1]
				}
			}
			continue
		}
		if dateStr == "" {
			continue
		}

		switch {
		case topBulletRe.MatchString(line):
			flush()
			inList = true
			current = &entry{date: date, lines: []string{topBulletRe.ReplaceAllString(line, "")}}
		case !inList:
		case strings.TrimSpace(line) == "" || strings.HasPrefix(line, " "):
			current.lines = append(current.lines, line)
		default:
			flush()
			inList = false
		}
	}
	flush()

	return entries
}

func buildItems(entries []entry) []*feeds.Item {
	var items []*feeds.Item
	seen := map[string]int{}
	for _, e := range entries {
		text := toPlainText(e.lines)
		if text == "" {
			continue
		}
		day := e.date.Format("2006-01-02")
		key := parser.TextKey(text)
		seen[day+" "+key]++

		items = append(items, &feeds.Item{
			Title:       parser.ExtractTitle(text),
			Content:     renderItem(e.lines),
			Description: text,
			Link:        &feeds.Link{Href: url},
			Created:     e.date,
			Id:          parser.GetGuid([]string{day, key, strconv.Itoa(seen[day+" "+key])}),
		})
	}
	return items
}

func (GeminiAPIChangelog) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed

	resp, err := parser.HttpGet(markdownUrl, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, parser.NewInternalError(fmt.Sprintf("unable to fetch the changelog page, status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed.Items = buildItems(parseEntries(string(body)))

	feed.Title = "Gemini API Changelog"
	feed.Description = "Release notes for the Gemini API"
	feed.Author = &feeds.Author{Name: "Google"}
	feed.Link = &feeds.Link{Href: url}

	parser.SortFeedEntries(&feed)

	return &feed, nil
}

func GeminiAPIChangelogParser() parser.Parser {
	return GeminiAPIChangelog{}
}
