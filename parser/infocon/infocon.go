package infocon

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (InfoCon) String() string {
	return "infocon"
}

func (InfoCon) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "url",
				Required: true,
				Type:     "string",
				Help:     "url of the infocon",
			},
		},
		Parser: InfoCon{},
	}
}

type InfoCon struct{}

func InfoConParser() parser.Parser {
	return InfoCon{}
}
func (InfoCon) Parse(options *parser.Options) (*feeds.Feed, error) {
	u := options.Get("url").(string)

	url, err := url.QueryUnescape(u)
	if err != nil {
		return nil, fmt.Errorf("error unescaping url")
	}
	resp, err := parser.HttpGet(url, nil)
	regexesIgnore := []*regexp.Regexp{
		regexp.MustCompile(`Thumbs\.db`),
		regexp.MustCompile(`.*\.jpg`),
		regexp.MustCompile(`.*\.webp`),
		regexp.MustCompile(`Parent directory/?`),
	}

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	table := doc.Find("table#list")
	trs := table.Find("tr")
	title := strings.TrimSpace(doc.Find("h1.breadcrumb").First().Text())

	feed := feeds.Feed{
		Title:       title,
		Description: title,
		Items:       []*feeds.Item{},
		Author:      &feeds.Author{Name: "InfoCon"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: url},
	}

	for i := range trs.Nodes {
		tr := trs.Eq(i)
		name := tr.Find("td.link").First().Text()
		href := tr.Find("td.link").First().Find("a").First().AttrOr("href", "")
		size := tr.Find("td.size").First().Text()
		date := tr.Find("td.date").First().Text()
		skip := href == "" || name == "" || size == "" || date == ""
		for _, regex := range regexesIgnore {
			if regex.MatchString(name) {
				skip = true
				continue
			}
		}
		if skip {
			continue
		}
		link := fmt.Sprintf("%s/%s", url, href)
		createdOn, err := time.Parse("2006 Jan 02 15:04", date)
		if err != nil {
			continue
		}

		newItem := feeds.Item{
			Title:       name,
			Content:     fmt.Sprintf("%s | %s | %s", name, size, date),
			Description: fmt.Sprintf("%s | %s | %s", name, size, date),
			Link:        &feeds.Link{Href: link},
			Created:     createdOn,
		}
		feed.Items = append(feed.Items, &newItem)

	}
	return &feed, nil
}
