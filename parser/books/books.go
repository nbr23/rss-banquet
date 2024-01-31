package books

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	unidecode "github.com/mozillazg/go-unidecode"
	"github.com/nbr23/atomic-banquet/parser"
)

type book struct {
	Title         string    `xml:"title"`
	Author        string    `xml:"author"`
	Language      string    `xml:"language"`
	PublishedDate time.Time `xml:"publishedDate"`
	VolumeLink    string    `xml:"volumeLink"`
}

func (b *book) NormalizedName() string {
	return strings.ToLower(unidecode.Unidecode(fmt.Sprintf("%s - %s - %s", b.Language, b.Title, b.Author)))
}

type volumeInfo struct {
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Language      string   `json:"language"`
	PublishedDate string   `json:"publishedDate"`
	VolumeLink    string   `json:"canonicalVolumeLink"`
}

type volume struct {
	VolumeInfo volumeInfo `json:"volumeInfo"`
	Id         string     `json:"id"`
}

type volumesReponse struct {
	Kind       string   `json:"kind"`
	Items      []volume `json:"items"`
	TotalItems int      `json:"totalItems"`
}

func containsLoose(l []string, s string) bool {
	ascii_s := strings.ToLower(unidecode.Unidecode(s))
	for _, e := range l {
		if strings.ToLower(unidecode.Unidecode(e)) == ascii_s {
			return true
		}
	}
	return false
}

func listBooksByForYear(booksList map[string]*book, author, language string, year int) {
	pageSize := 40
	for page := 0; ; page++ {
		url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=inauthor:%%22%s%%22+%d&langRestrict=%s&printType=books&orderBy=relevance&showPreorders=true&maxResults=%d&startIndex=%d", url.QueryEscape(author), year, language, pageSize, page*pageSize)

		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		var volRes *volumesReponse
		err = json.Unmarshal(resBody, &volRes)

		if err != nil {
			log.Fatal(err)
		}
		if len(volRes.Items) == 0 {
			break
		}
		for _, item := range volRes.Items {
			volumeAuthors := make([]string, len(item.VolumeInfo.Authors))
			for i, author := range item.VolumeInfo.Authors {
				volumeAuthors[i] = strings.ToLower(author)
			}
			if containsLoose(volumeAuthors, author) && item.VolumeInfo.Language == language {
				pubDate, err := dateparse.ParseStrict(item.VolumeInfo.PublishedDate)
				if err != nil {
					log.Println(err, item.VolumeInfo)
					continue
				}

				new_book := &book{
					strings.Title(item.VolumeInfo.Title),
					strings.Title(author),
					item.VolumeInfo.Language,
					pubDate,
					fmt.Sprintf("https://books.google.com/books/about/?hl=&id=%s", item.Id),
				}
				if _, ok := booksList[new_book.NormalizedName()]; !ok {
					booksList[new_book.NormalizedName()] = new_book
				}
			}
		}

		if volRes.TotalItems < pageSize*(1+page) {
			break
		}
	}
}

func (Books) Parse(options map[string]any) (*feeds.Feed, error) {
	var feed feeds.Feed

	author := options["author"].(string)
	language := parser.DefaultedGet(options, "language", "en")

	year_min := time.Now().Year() - 1
	year_max := time.Now().Year() + 1

	booksToSort := make(map[string]*book)
	for year := year_min; year <= year_max; year++ {
		listBooksByForYear(booksToSort, author, language, year)
	}
	for _, book := range booksToSort {
		if book.PublishedDate.Year() < year_min || book.PublishedDate.Year() > year_max {
			continue
		}
		item := &feeds.Item{
			Title:       fmt.Sprintf("%s - %s", book.Title, book.Author),
			Description: fmt.Sprintf("%s by %s published in %d", book.Title, book.Author, book.PublishedDate.Year()),
			Link:        &feeds.Link{Href: book.VolumeLink},
			Id:          fmt.Sprintf("%s%s", book.PublishedDate.Format(time.RFC3339), book.VolumeLink),
			Created:     book.PublishedDate,
			Updated:     book.PublishedDate,
		}
		feed.Items = append(feed.Items, item)
	}
	feed.Title = fmt.Sprintf("%s's books - %s", strings.Title(author), language)
	feed.Description = fmt.Sprintf("%s's books - %s", strings.Title(author), language)

	return &feed, nil
}

func (Books) Route(g *gin.Engine) gin.IRoutes {
	return g.GET("/books/:author", func(c *gin.Context) {
		feed, err := Books{}.Parse(map[string]any{
			"author":   c.Param("author"),
			"language": c.Query("language"),
		})
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		parser.ServeFeed(c, feed)
	})
}

func (Books) Help() string {
	return "\toptions:\n" +
		"\t - author\n" +
		"\t - language (default: en)\n"
}

type Books struct{}

func BooksParser() parser.Parser {
	return Books{}
}
