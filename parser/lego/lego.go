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

type legoItem struct {
	Name        string `json:"name"`
	ProductCode string `json:"productCode"`
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

func getLegoProductUrl(id string) string {
	return "https://www.lego.com/en-us/product/" + id
}

type Lego struct{}

func LegoParser() parser.Parser {
	return Lego{}
}

func (Lego) Help() string {
	return "\toptions:\n" +
		"\t - category: string (default: 'new', values: ['coming-soon', 'new'])\n"
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

func guid(item *legoItem) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(item.ProductCode, item.Name, item.Variant.Sku))))
}

func feedAdapter(l *legoFeed, options map[string]any) (*feeds.Feed, error) {
	feed := feeds.Feed{
		Title:       parser.DefaultedGet(options, "title", "Lego"),
		Description: parser.DefaultedGet(options, "description", "Lego Products"),
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
			Link:        &feeds.Link{Href: getLegoProductUrl(item.ProductCode)},
			Id:          guid(&item),
		}
		feed.Items = append(feed.Items, &newItem)
	}

	return &feed, nil

}

func getSlug(options map[string]any) string {
	category := parser.DefaultedGet(options, "category", "new")
	switch category {
	case "new":
		return "/categories/new-sets-and-products"
	case "coming-soon":
		return "/categories/coming-soon"
	default:
		return "new"
	}
}

func (Lego) Parse(options map[string]any) (*feeds.Feed, error) {
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
