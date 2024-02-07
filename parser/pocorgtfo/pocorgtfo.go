package pocorgtfo

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"
)

func guid(ss []string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(ss))))
}

func (PoCOrGTFO) Parse(options map[string]any) (*feeds.Feed, error) {
	const url = "https://www.alchemistowl.org/pocorgtfo/"
	var feed feeds.Feed
	pubRegex := regexp.MustCompile(`(?i)^(PoC\|\|GTFO 0x[0-9a-fA-F]{2})`)
	dateRegex := regexp.MustCompile(`(?i)^PoC||GTFO 0x[0-9a-fA-F]{2}, ([^,]+),`)

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
			date := dateRegex.FindStringSubmatch(item.Title)
			if len(date) < 1 {
				fmt.Println("unable to parse date", item.Title)
				return
			}
			item.Created, err = time.Parse("January 2, 2006", date[0])
			if err == nil {
				fmt.Println("unable to parse date", item.Title)
				return
			}
			feed.Items = append(feed.Items, &item)
		}
	})

	feed.Title = parser.DefaultedGet(options, "title", "PoC || GTFO")
	feed.Description = parser.DefaultedGet(options, "description", "PoC || GTFO Publications")
	feed.Author = &feeds.Author{
		Name: "PoC || GTFO",
	}
	feed.Link = &feeds.Link{Href: url}

	return &feed, nil
}

func (PoCOrGTFO) Route(g *gin.Engine) gin.IRoutes {
	return g.GET("/pocorgtfo", func(c *gin.Context) {
		feed, err := PoCOrGTFO{}.Parse(map[string]any{})
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		parser.ServeFeed(c, feed)
	})
}

func (PoCOrGTFO) Help() string {
	return ""
}

type PoCOrGTFO struct{}

func PoCOrGTFOParser() parser.Parser {
	return PoCOrGTFO{}
}
