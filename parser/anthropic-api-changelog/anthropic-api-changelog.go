package anthropicapichangelog

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const url = "https://docs.anthropic.com/en/release-notes/api"

type AnthropicAPIChangelog struct{}

func (AnthropicAPIChangelog) String() string {
	return "anthropic-api-changelog"
}

func (AnthropicAPIChangelog) GetOptions() parser.Options {
	return parser.Options{}
}

var ordinalRegex = regexp.MustCompile(`(\d+)(st|nd|rd|th)`)

func normalizeDate(s string) string {
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
		dateStr := strings.TrimSpace(h3.Text())
		date, err := time.Parse("January 2, 2006", normalizeDate(dateStr))
		if err != nil {
			return
		}

		ul := h3.NextAllFiltered("ul").First()
		if ul.Length() == 0 {
			return
		}

		ul.Find("li").Each(func(j int, li *goquery.Selection) {
			text := strings.TrimSpace(li.Text())
			if text == "" {
				return
			}

			title := extractTitle(text)

			htmlContent, err := li.Html()
			if err != nil {
				htmlContent = text
			}

			item := feeds.Item{
				Title:       title,
				Content:     htmlContent,
				Description: text,
				Link:        &feeds.Link{Href: url},
				Created:     date,
				Id:          parser.GetGuid([]string{dateStr, title}),
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
