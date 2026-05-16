package geminiapichangelog

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const url = "https://ai.google.dev/gemini-api/docs/changelog"

type GeminiAPIChangelog struct{}

func (GeminiAPIChangelog) String() string {
	return "gemini-api-changelog"
}

func (GeminiAPIChangelog) GetOptions() parser.Options {
	return parser.Options{}
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

func (GeminiAPIChangelog) Parse(options *parser.Options) (*feeds.Feed, error) {
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

	doc.Find("h2").Each(func(i int, h2 *goquery.Selection) {
		dateStr := strings.TrimSpace(h2.Text())
		date, err := time.Parse("January 2, 2006", dateStr)
		if err != nil {
			return
		}

		ul := h2.NextAllFiltered("ul").First()
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
