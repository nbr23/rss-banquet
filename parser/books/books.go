package books

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gorilla/feeds"
	unidecode "github.com/mozillazg/go-unidecode"
	"github.com/nbr23/rss-banquet/parser"
)

func (Books) String() string {
	return "books"
}

func (Books) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			&parser.Option{
				Flag:     "author",
				Required: true,
				Type:     "string",
				Help:     "author of the books",
				Value:    "",
			},
			&parser.Option{
				Flag:     "language",
				Required: false,
				Type:     "string",
				Help:     "language of the books",
				Default:  "en",
				Value:    "",
			},
		},
		Parser: Books{},
	}
}

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

func (Books) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed

	author := options.Get("author").(string)
	language := options.Get("language").(string)

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

		published := book.PublishedDate.Before(time.Now())
		item := &feeds.Item{
			Link:    &feeds.Link{Href: book.VolumeLink},
			Id:      fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s%s%s%v", book.Title, book.Author, book.Language, published)))),
			Created: book.PublishedDate,
			Updated: book.PublishedDate,
		}
		if published {
			item.Title = fmt.Sprintf("[PUBLISHED] %s - %s", book.Title, book.Author)
			item.Description = fmt.Sprintf("%s by %s published on %s", book.Title, book.Author, book.PublishedDate.Format("2006-01-02"))
		} else {
			item.Title = fmt.Sprintf("[ANNOUNCED] %s - %s - %s", book.Title, book.Author, book.Language)
			item.Description = fmt.Sprintf("%s by %s announced for %s", book.Title, book.Author, book.PublishedDate.Format("2006-01-02"))
		}
		feed.Items = append(feed.Items, item)
	}
	feed.Title = fmt.Sprintf("%s's books - %s", strings.Title(author), language)
	feed.Description = fmt.Sprintf("%s's books - %s", strings.Title(author), language)

	return &feed, nil
}

type Books struct{}

func BooksParser() parser.Parser {
	return Books{}
}
