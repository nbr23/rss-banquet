package anthropicapichangelog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const url = "https://platform.claude.com/docs/en/release-notes/overview"

type AnthropicAPIChangelog struct{}

func (AnthropicAPIChangelog) String() string {
	return "anthropic-api-changelog"
}

func (AnthropicAPIChangelog) GetOptions() parser.Options {
	return parser.Options{}
}

var ordinalRegex = regexp.MustCompile(`(\d+)(st|nd|rd|th)`)

func normalizeDate(s string) string {
	s = strings.TrimFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	return ordinalRegex.ReplaceAllString(s, "$1")
}

func extractTitle(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if idx := strings.Index(text, ". "); idx > 0 && idx < 200 {
		return text[:idx]
	}
	if len(text) > 120 {
		if idx := strings.LastIndex(text[:120], " "); idx > 0 {
			return text[:idx] + "..."
		}
		return text[:120] + "..."
	}
	return text
}

const idKeyWords = 6

func idKey(text string) string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(words) > idKeyWords {
		words = words[:idKeyWords]
	}
	return strings.Join(words, " ")
}

func (AnthropicAPIChangelog) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed

	resp, err := parser.HttpGet(url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, parser.NewInternalError(fmt.Sprintf("unable to fetch the changelog page, status code: %d", resp.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	doc.Find("h3").Each(func(i int, h3 *goquery.Selection) {
		dateStr := normalizeDate(h3.Text())
		date, err := time.Parse("January 2, 2006", dateStr)
		if err != nil {
			return
		}

		link := url
		if anchor, ok := h3.Attr("id"); ok {
			link = fmt.Sprintf("%s#%s", url, anchor)
		}

		seen := map[string]int{}
		h3.NextUntil("h2, h3").Filter("ul").ChildrenFiltered("li").Each(func(j int, li *goquery.Selection) {
			text := strings.TrimSpace(li.Text())
			if text == "" {
				return
			}

			title := extractTitle(text)

			key := idKey(text)
			seen[key]++

			htmlContent, err := li.Html()
			if err != nil {
				htmlContent = text
			}

			item := feeds.Item{
				Title:       title,
				Content:     htmlContent,
				Description: text,
				Link:        &feeds.Link{Href: link},
				Created:     date,
				Id:          parser.GetGuid([]string{dateStr, key, strconv.Itoa(seen[key])}),
			}
			feed.Items = append(feed.Items, &item)
		})
	})

	feed.Title = "Anthropic API Changelog"
	feed.Description = "Release notes for the Claude API"
	feed.Author = &feeds.Author{Name: "Anthropic"}
	feed.Link = &feeds.Link{Href: url}

	parser.SortFeedEntries(&feed)

	return &feed, nil
}

func AnthropicAPIChangelogParser() parser.Parser {
	return AnthropicAPIChangelog{}
}
