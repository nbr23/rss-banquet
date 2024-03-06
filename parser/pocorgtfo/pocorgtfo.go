package pocorgtfo

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"
)

func (PoCOrGTFO) String() string {
	return "pocorgtfo"
}

func (PoCOrGTFO) GetOptions() parser.Options {
	return parser.Options{}
}

func guid(ss []string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(ss))))
}

func (PoCOrGTFO) Parse(options *parser.Options) (*feeds.Feed, error) {
	const url = "https://www.alchemistowl.org/pocorgtfo/"
	var feed feeds.Feed
	pubRegex := regexp.MustCompile(`(?i)^(PoC\|\|GTFO 0x[0-9a-fA-F]{2})`)
	dateRegex := regexp.MustCompile(`(?i)^PoC\|\|GTFO 0x[0-9a-fA-F]{2}, ([^,]+),`)

	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to fetch the update page, status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a").First()
		pubMatches := pubRegex.FindStringSubmatch(link.Text())
		if len(pubMatches) > 0 && len(pubMatches[0]) >= 1 {
			var item feeds.Item
			item.Title = pubMatches[0]
			item.Description = s.Text()
			item.Link = &feeds.Link{Href: fmt.Sprintf("%s%s", url, link.AttrOr("href", ""))}
			item.Id = guid([]string{item.Title})
			date := dateRegex.FindStringSubmatch(s.Text())
			if len(date) < 2 || len(date[1]) <= 1 {
				fmt.Println("unable to find date", s.Text())
				return
			}
			item.Created, err = time.Parse("January 2006", date[1])
			if err != nil {
				fmt.Println("unable to parse date", s.Text())
				return
			}
			feed.Items = append(feed.Items, &item)
		}
	})

	feed.Title = "PoC || GTFO"
	feed.Description = "PoC || GTFO Publications"
	feed.Author = &feeds.Author{
		Name: "PoC || GTFO",
	}
	feed.Link = &feeds.Link{Href: url}

	return &feed, nil
}

type PoCOrGTFO struct{}

func PoCOrGTFOParser() parser.Parser {
	return PoCOrGTFO{}
}
