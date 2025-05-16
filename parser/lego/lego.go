package lego

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (Lego) String() string {
	return "lego"
}

func (Lego) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "category",
				Required: false,
				Type:     "string",
				Help:     "category of the lego products (new, coming-soon)",
				Default:  "new",
			},
		},
		Parser: Lego{},
	}
}

type legoItem struct {
	Name             string
	ProductCode      string
	ProductUrl       string
	Price            string
	AgeRange         string
	PieceCount       string
	AvailabilityText string
	ImgUrl           string
}

func getLegoProductUrl(l *legoItem) string {
	if l.ProductUrl != "" {
		return l.ProductUrl
	}
	return "https://www.lego.com/en-us/product/" + l.ProductCode
}

type Lego struct{}

func LegoParser() parser.Parser {
	return Lego{}
}

func buildItemTitle(item *legoItem) string {
	title := fmt.Sprintf("%s - %s", item.ProductCode, item.Name)
	available := item.AvailabilityText != "Coming Soon"
	if available {
		title = fmt.Sprintf("[NEW] %s", title)
	} else {
		title = fmt.Sprintf("[COMING SOON] %s", title)
	}
	if item.AvailabilityText != "" {
		title = fmt.Sprintf("%s - %s", title, item.AvailabilityText)
	}
	if item.Price != "" {
		title = fmt.Sprintf("%s - %s", title, item.Price)
	}
	if item.PieceCount != "" {
		title = fmt.Sprintf("%s - %s pieces", title, item.PieceCount)
	}
	if item.AgeRange != "" {
		title = fmt.Sprintf("%s %s", title, item.AgeRange)
	}
	return title
}

func buildItemContent(item *legoItem, isHtml bool) string {
	separator := "\n"
	if isHtml {
		separator = "<br/>"
	}
	description := fmt.Sprintf("%s - %s", item.ProductCode, item.Name)
	available := item.AvailabilityText != "Coming Soon"
	if available {
		description = fmt.Sprintf("%s (New)", description)
	} else {
		description = fmt.Sprintf("%s (Coming Soon)", description)
	}
	if item.Price != "" {
		description = fmt.Sprintf("%s - %s", description, item.Price)
	}
	if item.PieceCount != "" {
		description = fmt.Sprintf("%s - %s pieces", description, item.PieceCount)
	}
	if item.AgeRange != "" {
		description = fmt.Sprintf("%s %s", description, item.AgeRange)
	}
	if item.AvailabilityText != "" {
		description = fmt.Sprintf("%s%s%s", description, separator, item.AvailabilityText)
	}
	return description
}

func guid(item *legoItem, f feeds.Feed) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(f.Link.Href, item.ProductCode, item.Name))))
}

func getUrl(options *parser.Options) string {
	return fmt.Sprintf("https://www.lego.com/en-us%s", getSlug(options))
}

func feedAdapter(items []legoItem, options *parser.Options) (*feeds.Feed, error) {
	feed := feeds.Feed{
		Title:       fmt.Sprintf("Lego %s", getSlug(options)),
		Description: fmt.Sprintf("Lego %s Products", getSlug(options)),
		Items:       []*feeds.Item{},
		Author:      &feeds.Author{Name: "lego"},
		Created:     time.Now(),
		Link: &feeds.Link{
			Href: getUrl(options),
		},
	}

	for _, item := range items {
		if item.ProductCode == "" && item.Name == "" {
			log.Warn().Msgf("Skipping item with empty product code and name %v", item)
			continue
		}
		newItem := feeds.Item{
			Title:       buildItemTitle(&item),
			Content:     buildItemContent(&item, true),
			Description: buildItemContent(&item, false),
			Link:        &feeds.Link{Href: getLegoProductUrl(&item)},
			Id:          guid(&item, feed),
		}
		imgExt := parser.GetFileTypeFromUrl(item.ImgUrl)
		if !parser.IsImageType(imgExt) {
			imgExt = "png"
		}
		newItem.Enclosure = &feeds.Enclosure{
			Url:    item.ImgUrl,
			Type:   "image/" + imgExt,
			Length: "0",
		}
		feed.Items = append(feed.Items, &newItem)
	}

	return &feed, nil

}

func getSlug(options *parser.Options) string {
	category := options.Get("category").(string)
	switch category {
	case "new":
		return "/categories/new-sets-and-products"
	case "coming-soon":
		return "/categories/coming-soon"
	default:
		return "/categories/new-sets-and-products"
	}
}

func (Lego) Parse(options *parser.Options) (*feeds.Feed, error) {
	resp, err := parser.HttpGet(getUrl(options), nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to fetch the product page, status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	products := []legoItem{}
	doc.Find("li[data-test=product-item]").Each(func(i int, s *goquery.Selection) {
		if s.Children().Length() == 0 {
			return
		}
		if s.Find("div[data-test=product-listing-disruptor-static]").Length() > 0 {
			return
		}
		l := legoItem{}
		l.Name = s.Find("a[data-test=product-leaf-title]").First().Text()
		l.ProductUrl = s.Find("a[data-test=product-leaf-title]").First().AttrOr("href", "")
		if !strings.HasPrefix(l.ProductUrl, "http") {
			l.ProductUrl = "https://lego.com" + l.ProductUrl
		}
		l.ProductCode = s.Find("article[data-test=product-leaf]").First().AttrOr("data-test-key", "")
		l.Price = s.Find("span[data-test=product-leaf-price]").First().Text()
		l.AgeRange = s.Find("span[data-test=product-leaf-age-range-label]").First().Text()
		l.PieceCount = s.Find("span[data-test=product-leaf-piece-count-label]").First().Text()
		l.AvailabilityText = s.Find("div[data-test=product-leaf-action-row]").First().Find("button[data-test^=add-to-cart]").Text()
		if l.AvailabilityText == "" {
			l.AvailabilityText = s.Find("div[data-test=product-leaf-action-row]").First().Find("a[data-test$=-coming-soon]").Text()
		}
		if l.AvailabilityText == "Add to Bag" {
			l.AvailabilityText = "Available now"
		}
		if l.AvailabilityText == "Coming Soon" && options.Get("category") == "new" {
			return
		}
		l.ImgUrl = strings.Split(s.Find("ul[data-test=product-leaf-image-wrapper]").First().Find("source").First().AttrOr("srcset", ""), " ")[0]
		products = append(products, l)
	})

	return feedAdapter(products, options)
}
