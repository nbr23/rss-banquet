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

var ordinalRegex = regexp.MustCompile(`(\d+)(st|nd|rd|th)`)

func normalizeDate(s string) string {
	return ordinalRegex.ReplaceAllString(s, "$1")
}

func cleanText(s string) string {
	return strings.Join(strings.Fields(s), " ")
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

	doc.Find("div.intercom-interblocks-callout").Each(func(i int, callout *goquery.Selection) {
		h1 := callout.Find("h1").First()
		dateStr := cleanText(h1.Text())
		if dateStr == "" {
			return
		}
		date, err := time.Parse("January 2, 2006", normalizeDate(dateStr))
		if err != nil {
			return
		}

		var currentTitle string
		var currentHTML strings.Builder
		var currentText strings.Builder

		flush := func() {
			if currentTitle == "" {
				return
			}
			content := strings.TrimSpace(currentHTML.String())
			text := strings.TrimSpace(currentText.String())
			if content == "" {
				content = text
			}
			item := feeds.Item{
				Title:       currentTitle,
				Content:     content,
				Description: text,
				Link:        &feeds.Link{Href: url},
				Created:     date,
				Id:          parser.GetGuid([]string{dateStr, currentTitle}),
			}
			feed.Items = append(feed.Items, &item)
		}

		callout.Children().Each(func(j int, child *goquery.Selection) {
			tag := goquery.NodeName(child)
			switch tag {
			case "h1":
				return
			case "h2":
				flush()
				currentTitle = cleanText(child.Text())
				currentHTML.Reset()
				currentText.Reset()
			default:
				if currentTitle == "" {
					return
				}
				html, err := goquery.OuterHtml(child)
				if err == nil {
					currentHTML.WriteString(html)
				}
				text := strings.TrimSpace(child.Text())
				if text != "" {
					if currentText.Len() > 0 {
						currentText.WriteString("\n\n")
					}
					currentText.WriteString(text)
				}
			}
		})
		flush()
	})

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
