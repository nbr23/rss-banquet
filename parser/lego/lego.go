package lego

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"
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
				Value:    "",
			},
		},
		Parser: Lego{},
	}
}

type legoItem struct {
	Name        string `json:"name"`
	ProductCode string `json:"productCode"`
	OverrideUrl string `json:"overrideUrl"`
	BaseImgUrl  string `json:"baseImgUrl"`
	Variant     struct {
		Sku        string `json:"sku"`
		Id         string `json:"id"`
		Attributes struct {
			AgeRange         string `json:"ageRange"`
			PieceCount       int    `json:"pieceCount"`
			IsNew            bool   `json:"isNew"`
			OnSale           bool   `json:"onSale"`
			AvailabilityText string `json:"availabilityText"`
		} `json:"attributes"`
		Price struct {
			FormattedAmount string `json:"formattedAmount"`
		} `json:"price"`
		ListPrice struct {
			FormattedAmount string `json:"formattedAmount"`
		} `json:"listPrice"`
	} `json:"variant"`
}

type legoFeed struct {
	Data struct {
		ContentPage struct {
			ContentBody []struct {
				Section struct {
					Products struct {
						Results []legoItem `json:"results"`
					} `json:"products"`
				} `json:"section"`
			} `json:"contentBody"`
		} `json:"contentPage"`
	} `json:"data"`
}

func getLegoProductUrl(l *legoItem) string {
	if l.OverrideUrl != "" {
		return l.OverrideUrl
	}
	return "https://www.lego.com/en-us/product/" + l.ProductCode
}

type Lego struct{}

func LegoParser() parser.Parser {
	return Lego{}
}

func getLegoItemsFromFeed(feed *legoFeed) []legoItem {
	items := []legoItem{}
	for _, item := range feed.Data.ContentPage.ContentBody {
		items = append(items, item.Section.Products.Results...)
	}
	return items
}

func buildItemTitle(item *legoItem) string {
	title := fmt.Sprintf("%s - %s", item.ProductCode, item.Name)
	if item.Variant.Attributes.OnSale {
		title = fmt.Sprintf("[SALE] %s", title)
	}
	if item.Variant.Attributes.IsNew {
		title = fmt.Sprintf("[NEW] %s", title)
	}
	if item.Variant.Attributes.AvailabilityText != "" {
		title = fmt.Sprintf("%s - %s", title, item.Variant.Attributes.AvailabilityText)
	}
	if item.Variant.Price.FormattedAmount != "" {
		title = fmt.Sprintf("%s - %s", title, item.Variant.Price.FormattedAmount)
	}
	if item.Variant.Attributes.PieceCount != 0 {
		title = fmt.Sprintf("%s - %d pieces", title, item.Variant.Attributes.PieceCount)
	}
	if item.Variant.Attributes.AgeRange != "" {
		title = fmt.Sprintf("%s %s", title, item.Variant.Attributes.AgeRange)
	}
	return title
}

func buildItemDescription(item *legoItem) string {
	description := fmt.Sprintf("%s - %s", item.ProductCode, item.Name)
	if item.Variant.Attributes.IsNew {
		description = fmt.Sprintf("%s (New)", description)
	}
	if item.Variant.Price.FormattedAmount != "" {
		description = fmt.Sprintf("%s - %s", description, item.Variant.Price.FormattedAmount)
	}
	if item.Variant.Attributes.PieceCount != 0 {
		description = fmt.Sprintf("%s - %d pieces", description, item.Variant.Attributes.PieceCount)
	}
	if item.Variant.Attributes.AgeRange != "" {
		description = fmt.Sprintf("%s %s", description, item.Variant.Attributes.AgeRange)
	}
	if item.Variant.Attributes.AvailabilityText != "" {
		description = fmt.Sprintf("%s<br/>%s", description, item.Variant.Attributes.AvailabilityText)
	}
	if item.BaseImgUrl != "" {
		description = fmt.Sprintf("%s<br/><img src=\"%s\"/ alt=\"%s\">", description, item.BaseImgUrl, item.Name)
	}
	return description
}

func guid(item *legoItem, f feeds.Feed) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(f.Link.Href, item.ProductCode, item.Name, item.Variant.Sku))))
}

func feedAdapter(l *legoFeed, options *parser.Options) (*feeds.Feed, error) {
	feed := feeds.Feed{
		Title:       fmt.Sprintf("Lego %s", getSlug(options)),
		Description: fmt.Sprintf("Lego %s Products", getSlug(options)),
		Items:       []*feeds.Item{},
		Author:      &feeds.Author{Name: "lego"},
		Created:     time.Now(),
		Link: &feeds.Link{
			Href: fmt.Sprintf("https://www.lego.com/en-us%s", getSlug(options)),
		},
	}

	for _, item := range getLegoItemsFromFeed(l) {
		newItem := feeds.Item{
			Title:       buildItemTitle(&item),
			Description: buildItemDescription(&item),
			Link:        &feeds.Link{Href: getLegoProductUrl(&item)},
			Id:          guid(&item, feed),
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
		return "new"
	}
}

func (Lego) Parse(options *parser.Options) (*feeds.Feed, error) {
	resp, err := legoFeedQuery(options)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var feed legoFeed

	if err := json.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	return feedAdapter(&feed, options)
}
