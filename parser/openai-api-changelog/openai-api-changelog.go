package openaiapichangelog

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const url = "https://developers.openai.com/changelog/"

type OpenAIAPIChangelog struct{}

func (OpenAIAPIChangelog) String() string {
	return "openai-api-changelog"
}

func (OpenAIAPIChangelog) GetOptions() parser.Options {
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

func sectionYear(sectionStr string) string {
	parts := strings.SplitN(sectionStr, ",", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (OpenAIAPIChangelog) Parse(options *parser.Options) (*feeds.Feed, error) {
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

	doc.Find(`h3[class*="_ChangelogSectionTitle_"]`).Each(func(i int, h3 *goquery.Selection) {
		sectionStr := strings.TrimSpace(h3.Text())
		year := sectionYear(sectionStr)
		if year == "" {
			return
		}

		section := h3.Parent()
		section.Find(`div[class*="_ChangelogMarkdown_"]`).Each(func(j int, md *goquery.Selection) {
			entry := md.Closest("div.grid")
			if entry.Length() == 0 {
				return
			}

			dayStr := strings.TrimSpace(entry.Find(`div[data-variant="outline"]`).First().Text())
			if dayStr == "" {
				return
			}

			fullDateStr := fmt.Sprintf("%s, %s", dayStr, year)
			date, err := time.Parse("January 2, 2006", normalizeDate(fullDateStr))
			if err != nil {
				return
			}

			var typeStr string
			var tags []string
			entry.Find(`div[data-variant="soft"]`).Each(func(k int, b *goquery.Selection) {
				t := strings.TrimSpace(b.Text())
				if t == "" {
					return
				}
				if typeStr == "" {
					typeStr = t
				} else {
					tags = append(tags, t)
				}
			})

			bodyText := strings.TrimSpace(md.Text())
			if bodyText == "" {
				return
			}

			htmlContent, err := md.Html()
			if err != nil {
				htmlContent = bodyText
			}

			title := extractTitle(bodyText)
			if typeStr != "" {
				title = fmt.Sprintf("%s: %s", typeStr, title)
			}
			if len(tags) > 0 {
				title = fmt.Sprintf("[%s] %s", strings.Join(tags, ", "), title)
			}

			item := feeds.Item{
				Title:       title,
				Content:     htmlContent,
				Description: bodyText,
				Link:        &feeds.Link{Href: url},
				Created:     date,
				Id:          parser.GetGuid([]string{fullDateStr, title}),
			}
			feed.Items = append(feed.Items, &item)
		})
	})

	feed.Title = "OpenAI API Changelog"
	feed.Description = "Release notes for the OpenAI developer platform"
	feed.Author = &feeds.Author{Name: "OpenAI"}
	feed.Link = &feeds.Link{Href: url}

	parser.SortFeedEntries(&feed)

	return &feed, nil
}

func OpenAIAPIChangelogParser() parser.Parser {
	return OpenAIAPIChangelog{}
}
