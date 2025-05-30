package nytimes

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

type NYTimes struct{}

func NYTimesParser() parser.Parser {
	return NYTimes{}
}

func (NYTimes) String() string {
	return "nytimes"
}

func (NYTimes) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "author",
				Required: false,
				Type:     "string",
				Help:     "author of the articles to fetch",
			},
		},
		Parser: NYTimes{},
	}
}

func (NYTimes) Parse(options *parser.Options) (*feeds.Feed, error) {
	author, ok := options.Get("author").(string)
	if !ok || author == "" {
		return nil, fmt.Errorf("author is required")
	}
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("Articles by %s - The New York Times", author),
		Link:        &feeds.Link{Href: fmt.Sprintf("https://www.nytimes.com/by/%s#latest", author)},
		Description: fmt.Sprintf("Latest articles by %s on The New York Times", author),
		Author:      &feeds.Author{Name: author},
	}

	articles, err := getGraphQLResponse(author)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch articles: %w", err)
	}

	for _, edge := range articles {
		node := edge.Node
		if node.ID == "" || node.Headline.Default == "" || node.URL == "" {
			continue
		}
		title := node.Headline.Default
		link := node.URL
		description := node.Summary

		parsedTime, err := time.Parse(time.RFC3339, node.FirstPublished)

		if err != nil {
			continue
		}

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: link},
			Description: description,
			Created:     parsedTime,
		})
	}
	return feed, nil
}
