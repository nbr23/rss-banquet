package costco

import (
	"compress/gzip"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (Costco) String() string {
	return "costco"
}

func (Costco) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "url",
				Required: true,
				Type:     "string",
				Help:     "URL of the Costco page to scrape",
				IsPath:   true,
			},
		},
		Parser: Costco{},
	}
}

type Costco struct{}

func CostcoParser() parser.Parser {
	return Costco{}
}

func (Costco) Parse(options *parser.Options) (*feeds.Feed, error) {
	var skuPattern = regexp.MustCompile(`(?m)^[\s]*SKU: '([^']+)'`)
	var namePattern = regexp.MustCompile(`(?m)^[\s]*name: '([^']+)'`)
	var pricePattern = regexp.MustCompile(`(?m)^\s+priceTotal: (.+[^,]),?$`)
	var imagePattern = regexp.MustCompile(`(?m)^\s+productImageUrl: '([^']+)'`)

	url := options.Get("url").(string)[1:]
	if url == "" {
		return nil, fmt.Errorf("url is required")
	}

	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.4 Safari/605.1.15",
		"Accept-Encoding": "gzip, deflate, br, zstd",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	}

	resp, err := parser.HttpGet(url, map[string]any{
		"headers": headers,
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	zipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch page: %s", resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(zipReader)
	if err != nil {
		return nil, err
	}
	feed := &feeds.Feed{
		Title:       "Costco Products",
		Link:        &feeds.Link{Href: url},
		Description: "Latest products from Costco",
	}

	// parse title of page
	title := doc.Find("title").First().Text()
	if len(title) > 0 {
		feed.Title = title
		feed.Description = title
	}

	doc.Find(".product-tile-set").Each(func(i int, s *goquery.Selection) {
		script := s.Find("script[type='text/javascript']").First().Text()
		if len(script) == 0 {
			return
		}

		SKUMatches := skuPattern.FindStringSubmatch(script)
		if len(SKUMatches) < 2 || SKUMatches[1] == "" {
			return
		}
		SKU := SKUMatches[1]

		titleMatches := namePattern.FindStringSubmatch(script)
		if len(titleMatches) < 2 || titleMatches[1] == "" {
			return
		}
		title := titleMatches[1]

		priceMatches := pricePattern.FindStringSubmatch(script)
		if len(priceMatches) < 2 || priceMatches[1] == "" {
			return
		}
		price := priceMatches[1]

		imageMatches := imagePattern.FindStringSubmatch(script)
		if len(imageMatches) < 2 || imageMatches[1] == "" {
			return
		}
		imageUrl := imageMatches[1]

		productFeatures := s.Find("ul.product-features").First().Text()

		newItem := feeds.Item{
			Title:       fmt.Sprintf("%s - $%s", title, price),
			Description: productFeatures,
			Link:        &feeds.Link{Href: fmt.Sprintf("https://www.costco.com/s?dept=All&keyword=%s", SKU)},
			Id:          SKU,
		}
		if imageUrl != "" {
			newItem.Enclosure = &feeds.Enclosure{
				Url:    strings.ReplaceAll(imageUrl, "\\/", "/"),
				Type:   "image/jpeg",
				Length: "0",
			}
		}
		feed.Items = append(feed.Items, &newItem)
	})

	return feed, nil
}
