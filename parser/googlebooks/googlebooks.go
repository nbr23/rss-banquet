package googlebooks

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/araddon/dateparse"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
	"github.com/nbr23/rss-banquet/utils"
)

type Googlebooks struct{}

func (Googlebooks) String() string {
	return "books"
}

func GooglebooksParser() parser.Parser {
	return Googlebooks{}
}

type book struct {
	Title         string    `json:"title"`
	Authors       []string  `json:"authors"`
	Language      string    `json:"language"`
	PublishedDate time.Time `json:"publishedDate"`
	VolumeLink    string    `json:"canonicalVolumeLink"`
	CoverUrl      string
	Publisher     string
}

func getSearchUrl(author string, language string, year_min int, year_max int) string {
	u := "https://www.google.com/search?q="
	for _, word := range strings.Split(author, " ") {
		u += "inauthor:" + url.QueryEscape(word) + "+"
	}
	u += "&lr=lang_" + language
	u += fmt.Sprintf("&tbs=cdr:1,cd_min:Jan+1_2+%d,cd_max:Dec+31_2+%d,lr:lang_1%s,bkt:b,bkv:p&tbm=bks&source=lnt", year_min, year_max, language)
	return u
}

func generateSOCSCookie() string {
	currentDate := time.Now().Format("20060102")

	cookieContent := []byte{
		0x08, 0x02, 0x12, 0x35, 0x08, 0x02, 0x12, 0x2b,
	}
	cookieContent = append(cookieContent, []byte(fmt.Sprintf("boq_identityfrontenduiserver_%s.08_p0", currentDate))...)
	cookieContent = append(cookieContent, 0x1A, 0x02, 'e', 'n', 0x20, 0x02, 0x1A, 0x06, 0x08, 0x80, 0xD1, 0xB9)

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write(cookieContent)
	encoder.Close()

	socsCookie := strings.ReplaceAll(buf.String(), "\n", "")

	return socsCookie
}

func fetchGoogleBooksPage(u string) (*http.Response, error) {
	req, err := http.NewRequest(
		"GET",
		u,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("DNT", "1")
	req.Header.Set("Cookie", fmt.Sprintf("SOCS=%s", generateSOCSCookie()))
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getBookDetailFromHtml(id string) (*book, error) {
	bookUrl := fmt.Sprintf("https://books.google.com/books?id=%s&redir_esc=y", id)

	resp, err := fetchGoogleBooksPage(bookUrl)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to fetch the book page, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(body)
	bodyString = bodyString[strings.Index(bodyString, "</head>")+7:] + "</body>"

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
	if err != nil {
		return nil, err
	}

	book := book{}

	book.Title = doc.Find("h1[class='booktitle']").Text()
	book.CoverUrl = doc.Find("img[title='Front Cover']").First().AttrOr("src", "")

	book.Authors = []string{}
	bookinfo := doc.Find("div[class='bookinfo_sectionwrap']").First().Children().First()
	bookinfo.Each(func(i int, s *goquery.Selection) {
		book.Authors = append(book.Authors, s.Text())
	})
	bookinfo = bookinfo.Next()
	book.Publisher = bookinfo.Find("span").First().Text()
	publishedDate := bookinfo.Find("span").First().Next().Text()

	pubDate, err := dateparse.ParseStrict(publishedDate)
	if err != nil {
		return nil, err
	}
	book.PublishedDate = pubDate
	book.VolumeLink = bookUrl

	return &book, nil
}

func (Googlebooks) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed

	author := options.Get("author").(string)
	language := options.Get("language").(string)

	year_min := time.Now().Year() - 1
	year_max := time.Now().Year() + 1

	searchUrl := getSearchUrl(author, language, year_min, year_max)

	resp, err := fetchGoogleBooksPage(searchUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to fetch the book search page, status code: %d", resp.StatusCode)
	}

	authorQuery := ""
	for _, word := range strings.Split(author, " ") {
		authorQuery += url.QueryEscape("inauthor:") + url.QueryEscape(word) + "%20"
	}
	authorQuery = strings.TrimSuffix(authorQuery, "%20")
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	bookIds := []string{}

	doc.Find(fmt.Sprintf("div[data-async-context='query:%s']", authorQuery)).Each(func(i int, s *goquery.Selection) {
		s.Children().Each(func(i int, s *goquery.Selection) {
			s.Find("a").Each(func(i int, s *goquery.Selection) {
				bookUrlRe := regexp.MustCompile(`https:\/\/books\.google\.[a-z]+/books\?id=([^&]+)&`)
				if bookUrlRe.MatchString(s.AttrOr("href", "")) {
					bookId := bookUrlRe.FindStringSubmatch(s.AttrOr("href", ""))[1]
					bookIds = utils.InsertUnique[string](bookIds, bookId)
				}
			})
		})
	})

	for _, bookId := range bookIds {
		book, err := getBookDetailFromHtml(bookId)
		if err != nil {
			book, err = getBookDetailFromHtml(bookId)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		authorFound := false
		for _, bookAuthor := range book.Authors {
			if strings.Contains(strings.ToLower(bookAuthor), strings.ToLower(author)) {
				authorFound = true
				break
			}
		}
		if !authorFound {
			continue
		}

		if book.Language == "" {
			book.Language = language
		}
		published := book.PublishedDate.Before(time.Now())
		item := &feeds.Item{
			Link:    &feeds.Link{Href: book.VolumeLink},
			Id:      fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s%s%s%v", book.Title, strings.Join(book.Authors, ", "), book.Language, published)))),
			Created: book.PublishedDate,
			Updated: book.PublishedDate,
		}
		if published {
			item.Title = fmt.Sprintf("[PUBLISHED] %s - %s", book.Title, strings.Join(book.Authors, ", "))
			item.Content = fmt.Sprintf("%s by %s published on %s", book.Title, strings.Join(book.Authors, ", "), book.PublishedDate.Format("2006-01-02"))
		} else {
			item.Title = fmt.Sprintf("[ANNOUNCED] %s - %s - %s", book.Title, strings.Join(book.Authors, ", "), book.Language)
			item.Content = fmt.Sprintf("%s by %s announced for %s", book.Title, strings.Join(book.Authors, ", "), book.PublishedDate.Format("2006-01-02"))
		}
		item.Description = item.Content
		imgExt := parser.GetFileTypeFromUrl(book.CoverUrl)
		if !parser.IsImageType(imgExt) {
			imgExt = "png"
		}
		if book.CoverUrl != "" {
			item.Enclosure = &feeds.Enclosure{
				Url:    book.CoverUrl,
				Type:   "image/" + imgExt,
				Length: "0",
			}
		}
		feed.Items = append(feed.Items, item)

	}
	feed.Title = fmt.Sprintf("%s's books - %s", strings.Title(author), language)
	feed.Description = fmt.Sprintf("%s's books - %s", strings.Title(author), language)
	feed.Link = &feeds.Link{Href: getSearchUrl(author, language, year_min, year_max)}
	return &feed, nil
}

func (Googlebooks) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			&parser.Option{
				Flag:     "author",
				Required: false,
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
		Parser: Googlebooks{},
	}
}
