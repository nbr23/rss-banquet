package goodreads

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

type GoodReads struct{}

func (GoodReads) String() string {
	return "goodreads"
}

func GoodReadsParser() parser.Parser {
	return GoodReads{}
}

var EditionTypes = []string{
	"Paperback",
	"Hardcover",
	"Mass Market Paperback",
	"Kindle Edition",
	"Nook",
	"ebook",
	"Library Binding",
	"Audiobook",
	"Audio CD",
	"Audio Cassette",
	"Audible Audio",
	"CD-ROM",
	"MP3 CD",
	"Board book",
	"Leather Bound",
	"Unbound",
	"Spiral-bound",
	"Unknown Binding",
}

// Grabs rudimentary book details from the editions page
func getBookEditions(editionsUrl string) ([]*GRBook, error) {
	resp, err := parser.HttpGet(editionsUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	books := []*GRBook{}
	var pubRe = regexp.MustCompile(`Published[\s]+[A-Za-z\s0-9]*(\d{4})`)
	var expectedRe = regexp.MustCompile(`expected[\s]+publication[\s]+(\d{4})`)

	doc.Find("div[class='editionData']").Each(func(i int, s *goquery.Selection) {
		book := GRBook{}
		book.Link = s.Find("a.bookTitle").First().AttrOr("href", "")
		if book.Link == "" {
			return
		}
		book.Link = fmt.Sprintf("https://www.goodreads.com%s", book.Link)

		s.Find("div[class='dataRow']").Each(func(i int, s *goquery.Selection) {
			dataTitle := s.Find("div[class='dataTitle']").First().Text()
			if dataTitle != "" {
				if strings.Contains(dataTitle, "Edition language:") {
					book.Language = strings.TrimSpace(s.Find("div[class='dataValue']").First().Text())
					if book.Language == "" {
						book.Language = "English"
					}
				}
			} else {
				bookFormat := getBookFormatFromPageFormat(strings.Join(strings.Fields(s.Text()), " "))
				if bookFormat != "" {
					book.BookFormat = bookFormat
				}

				published := pubRe.MatchString(s.Text())
				if published {
					book.PublicationDate = pubRe.FindStringSubmatch(s.Text())[1]
				} else if expectedRe.MatchString(s.Text()) {
					book.PublicationDate = expectedRe.FindStringSubmatch(s.Text())[1]
				}
			}
		})
		books = append(books, &book)
	})
	return books, nil
}

func getBookFormatFromPageFormat(pageformat string) string {
	for _, t := range EditionTypes {
		if strings.Contains(pageformat, t) {
			return t
		}
	}
	return ""
}

func getBookDetails(bookLink string) (*GRBook, error) {
	resp, err := parser.HttpGet(bookLink)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	book := GRBook{}

	pubInfo := doc.Find("p[data-testid='publicationInfo']").First().Text()
	titleSection := doc.Find("div[class='BookPageTitleSection__title']").First()
	book.SubTitle = strings.Join(strings.Fields(titleSection.Find("h3").Text()), " ")
	book.Title = strings.Join(strings.Fields(doc.Find("h1[data-testid='bookTitle']").Text()), " ")
	book.Author = strings.Join(strings.Fields(doc.Find("div[class='BookPageMetadataSection__contributor']").Text()), " ")
	book.BookFormat = getBookFormatFromPageFormat(strings.Join(strings.Fields(doc.Find("p[data-testid='pagesFormat']").First().Text()), " "))
	book.Description = strings.Join(strings.Fields(doc.Find("div[class='BookPageMetadataSection__description']").First().Text()), " ")
	book.PublicationDate = pubInfo
	book.Link = bookLink
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "inLanguage") {
			var bookJson GRBookJson
			err := json.Unmarshal([]byte(s.Text()), &bookJson)
			if err != nil {
				log.Error().Msg(fmt.Sprintf("unable to parse book json: %s", err.Error()))
				return
			}
			book.Language = bookJson.InLanguage
			book.CoverUrl = bookJson.Image
		}
	})
	return &book, nil
}

func getAuthorBooksList(authorId string, bookLanguage string, yearMin int, bookFormats []string) (string, string, []GRBook, error) {
	url := fmt.Sprintf("https://www.goodreads.com/author/list/%s?utf8=%%E2%%9C%%93&sort=original_publication_year", authorId)
	books, title, err := getBooksList(url, bookLanguage, yearMin, bookFormats)
	return url, title, books, err
}

func getSeriesBooksList(seriesId string, bookLanguage string, yearMin int, bookFormats []string) (string, string, []GRBook, error) {
	url := fmt.Sprintf("https://www.goodreads.com/series/%s", seriesId)
	books, title, err := getBooksList(url, bookLanguage, yearMin, bookFormats)
	return url, title, books, err
}

func getBooksList(url string, bookLanguage string, yearMin int, bookFormats []string) ([]GRBook, string, error) {
	resp, err := parser.HttpGet(url)
	if err != nil {
		return nil, "", parser.NewInternalError("unable to fetch the page")
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", parser.NewInternalError("unable to parse the page")
	}

	title := doc.Find("h1").First().Text()

	if title == "" {
		return nil, "", parser.NewNotFoundError("book list page not found")
	}

	pubRe := regexp.MustCompile(`published[\s]+(\d{4})`)
	expectedRe := regexp.MustCompile(`expected[\s]+publication[\s]+(\d{4})`)

	books := []GRBook{}

	doc.Find("[itemtype='http://schema.org/Book']").Each(func(i int, s *goquery.Selection) {
		titleSection := s.Find("a[itemprop='url']").First()
		title := strings.Join(strings.Fields(titleSection.Text()), " ")
		bookLink := titleSection.AttrOr("href", "")
		if strings.HasPrefix(bookLink, "/") {
			bookLink = fmt.Sprintf("https://www.goodreads.com%s", bookLink)
		}
		published := pubRe.MatchString(s.Text())
		if !published && !expectedRe.MatchString(s.Text()) {
			log.Warn().Msg(fmt.Sprintf("Unexpected publishing info for %s: %s", title, s.Text()))
			return
		}
		var pubYear string
		if published {
			pubYear = pubRe.FindStringSubmatch(s.Text())[1]
		} else {
			pubYear = expectedRe.FindStringSubmatch(s.Text())[1]
		}
		if pubYear == "" {
			return
		}

		var editionsUrl string
		s.Find("a[href^='/work/editions/']").Each(func(i int, s *goquery.Selection) {
			editionsUrl = s.AttrOr("href", "")
		})
		if editionsUrl != "" {
			editionsUrl = fmt.Sprintf("https://www.goodreads.com%s", editionsUrl)

			editions, err := getBookEditions(editionsUrl)
			if err != nil {
				return
			}
			var earliestEdition *GRBook
			var earliestEditionYear int

			for _, e := range editions {
				fmt.Println("Edition:", e.Link, e.PublicationDate, e.BookFormat, e.Language)
				if e.Language != bookLanguage {
					continue
				}
				if e.BookFormat == "" {
					continue
				}
				if !isAcceptedBookFormat(bookFormats, e.BookFormat) {
					continue
				}

				if e.PublicationDate != "" {
					year, err := time.Parse("2006", e.PublicationDate)
					if err != nil {
						continue
					}
					if year.Year() <= earliestEditionYear || earliestEdition == nil {
						earliestEdition = e
						earliestEditionYear = year.Year()
					}
				}
			}

			if earliestEdition != nil {
				bookLink = earliestEdition.Link
			}
		}

		year, err := time.Parse("2006", pubYear)
		if err != nil || year.Year() < yearMin {
			return
		}
		book, err := getBookDetails(bookLink)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("unable to fetch book details: %s", err.Error()))
			return
		}
		if book.Language != bookLanguage {
			return
		}
		if !isAcceptedBookFormat(bookFormats, book.BookFormat) {
			return
		}
		books = append(books, *book)
	})

	return books, title, nil
}

func isAcceptedBookFormat(acceptedFormats []string, bookFormat string) bool {
	for _, f := range acceptedFormats {
		if strings.Contains(strings.ToLower(bookFormat), strings.ToLower(f)) {
			return true
		}
	}
	return false
}

type GRBookJson struct {
	Name          string `json:"name"`
	Image         string `json:"image"`
	BookFormat    string `json:"bookFormat"`
	NumberOfPages int    `json:"numberOfPages"`
	InLanguage    string `json:"inLanguage"`
	Isbn          string `json:"isbn"`
}

type GRBook struct {
	Title           string
	SubTitle        string
	PublicationDate string
	Link            string
	Author          string
	BookFormat      string
	Description     string
	Language        string
	CoverUrl        string
}

func getBookLanguage(bookLanguage string) (string, error) {
	tag, err := language.Parse(bookLanguage)
	if err != nil {
		return "", parser.NewNotFoundError("language not found")
	}
	return display.English.Languages().Name(tag), nil
}

func (GoodReads) Parse(options *parser.Options) (*feeds.Feed, error) {
	authorId := options.Get("authorId").(string)
	seriesId := options.Get("seriesId").(string)
	yearMin := options.Get("year-min").(int)
	bookLanguage := options.Get("language").(string)
	bookFormats := options.Get("bookFormats").([]string)

	if bookLanguage != "" {
		var err error
		bookLanguage, err = getBookLanguage(bookLanguage)
		if err != nil {
			return nil, err
		}
	}

	var books []GRBook
	var err error
	var url string
	var title string

	if authorId != "" {
		url, title, books, err = getAuthorBooksList(authorId, bookLanguage, yearMin, bookFormats)
	} else if seriesId != "" {
		url, title, books, err = getSeriesBooksList(seriesId, bookLanguage, yearMin, bookFormats)
	} else {
		return nil, parser.NewNotFoundError("authorId or seriesId required")
	}
	if err != nil {
		return nil, err
	}

	var feed feeds.Feed

	for _, book := range books {
		var item feeds.Item
		item.Title = fmt.Sprintf("%s - %s", book.Title, book.Author)
		item.Content = fmt.Sprintf("%s by %s published on %s", book.Title, book.Author, book.PublicationDate)
		item.Description = item.Content
		item.Link = &feeds.Link{Href: book.Link}
		item.Id = fmt.Sprintf("%s|%s", book.Link, book.PublicationDate)
		item.Created = time.Now()
		item.Updated = time.Now()
		feed.Items = append(feed.Items, &item)

		if book.CoverUrl != "" {
			imgExt := parser.GetFileTypeFromUrl(book.CoverUrl)
			if !parser.IsImageType(imgExt) {
				imgExt = "png"
			}
			item.Enclosure = &feeds.Enclosure{
				Url:    book.CoverUrl,
				Type:   "image/" + imgExt,
				Length: "0",
			}
		}
	}

	feed.Title = fmt.Sprintf("%s - %s", title, bookLanguage)
	feed.Description = feed.Title
	feed.Link = &feeds.Link{Href: url}

	return &feed, nil
}

func (GoodReads) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "authorId",
				Required: false,
				Type:     "string",
				Help:     "Goodreads author ID",
				Value:    "",
			},
			{
				Flag:     "seriesId",
				Required: false,
				Type:     "string",
				Help:     "Goodreads series ID",
				Value:    "",
			},
			{
				Flag:     "year-min",
				Required: false,
				Type:     "int",
				Help:     "minimum year of publication",
				Default:  fmt.Sprintf("%d", time.Now().Year()-1),
			},
			{
				Flag:     "language",
				Required: false,
				Type:     "string",
				Help:     "language of the book",
				Default:  "en",
			},
			{
				Flag:     "bookFormats",
				Required: false,
				Type:     "stringSlice",
				Help:     "seeked formats of the book (paperback, hardcover, ebook, audiobook, etc.)",
				Default:  "paperback,hardcover,kindle,ebook",
			},
		},
		Parser: GoodReads{},
	}
}
