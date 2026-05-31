package imdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const graphqlEndpoint = "https://caching.graphql.imdb.com/"

type IMDB struct{}

func (IMDB) String() string {
	return "imdb"
}

func IMDBParser() parser.Parser {
	return IMDB{}
}

type IMDBWork struct {
	TitleID     string
	Title       string
	Year        int
	ReleaseDate time.Time
	HasFullDate bool
	Category    string
	TitleType   string
	Role        string
	Link        string
	ImageURL    string
}

type gqlCredit struct {
	Title struct {
		ID        string `json:"id"`
		TitleText struct {
			Text string `json:"text"`
		} `json:"titleText"`
		ReleaseYear *struct {
			Year int `json:"year"`
		} `json:"releaseYear"`
		ReleaseDate *struct {
			Day   *int `json:"day"`
			Month *int `json:"month"`
			Year  *int `json:"year"`
		} `json:"releaseDate"`
		TitleType *struct {
			Text string `json:"text"`
		} `json:"titleType"`
		PrimaryImage *struct {
			URL string `json:"url"`
		} `json:"primaryImage"`
	} `json:"title"`
	Category *struct {
		Text string `json:"text"`
	} `json:"category"`
	Characters []struct {
		Name string `json:"name"`
	} `json:"characters"`
}

type gqlResponse struct {
	Data struct {
		Name struct {
			NameText struct {
				Text string `json:"text"`
			} `json:"nameText"`
			Credits struct {
				Edges []struct {
					Node gqlCredit `json:"node"`
				} `json:"edges"`
			} `json:"credits"`
		} `json:"name"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type httpPostFunc func(url, contentType string, body io.Reader) (*http.Response, error)

func buildQuery(artistId string, first int) []byte {
	gql := map[string]any{
		"query": `query NameCredits($id: ID!, $first: Int!) {
			name(id: $id) {
				nameText { text }
				credits(first: $first) {
					edges {
						node {
							title {
								id
								titleText { text }
								releaseYear { year }
								releaseDate { day month year }
								titleType { text }
								primaryImage { url }
							}
							category { text }
							... on Cast { characters { name } }
						}
					}
				}
			}
		}`,
		"variables": map[string]any{
			"id":    artistId,
			"first": first,
		},
	}
	b, _ := json.Marshal(gql)
	return b
}

func getArtistWorks(artistId string, first int, post httpPostFunc) ([]IMDBWork, string, error) {
	body := buildQuery(artistId, first)
	resp, err := post(graphqlEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, "", parser.NewInternalError(fmt.Sprintf("imdb graphql request failed: %s", err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", parser.NewInternalError("unable to read imdb graphql response")
	}

	var data gqlResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, "", parser.NewInternalError("unable to parse imdb graphql response")
	}
	if len(data.Errors) > 0 {
		return nil, "", parser.NewInternalError(fmt.Sprintf("imdb graphql error: %s", data.Errors[0].Message))
	}

	artistName := data.Data.Name.NameText.Text
	if artistName == "" {
		return nil, "", parser.NewNotFoundError("artist not found")
	}

	works := make([]IMDBWork, 0, len(data.Data.Name.Credits.Edges))
	for _, edge := range data.Data.Name.Credits.Edges {
		n := edge.Node
		if n.Title.ID == "" {
			continue
		}
		work := IMDBWork{
			TitleID: n.Title.ID,
			Title:   n.Title.TitleText.Text,
			Link:    fmt.Sprintf("https://www.imdb.com/title/%s/", n.Title.ID),
		}
		if n.Title.ReleaseYear != nil {
			work.Year = n.Title.ReleaseYear.Year
		}
		if n.Title.ReleaseDate != nil && n.Title.ReleaseDate.Year != nil {
			year := *n.Title.ReleaseDate.Year
			month := 1
			day := 1
			if n.Title.ReleaseDate.Month != nil {
				month = *n.Title.ReleaseDate.Month
			}
			if n.Title.ReleaseDate.Day != nil {
				day = *n.Title.ReleaseDate.Day
			}
			work.ReleaseDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			work.HasFullDate = n.Title.ReleaseDate.Month != nil && n.Title.ReleaseDate.Day != nil
		} else if work.Year > 0 {
			work.ReleaseDate = time.Date(work.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		}
		if n.Title.TitleType != nil {
			work.TitleType = n.Title.TitleType.Text
		}
		if n.Category != nil {
			work.Category = n.Category.Text
		}
		if n.Title.PrimaryImage != nil {
			work.ImageURL = n.Title.PrimaryImage.URL
		}
		if len(n.Characters) > 0 {
			names := make([]string, 0, len(n.Characters))
			for _, c := range n.Characters {
				if c.Name != "" {
					names = append(names, c.Name)
				}
			}
			work.Role = strings.Join(names, ", ")
		}
		works = append(works, work)
	}
	return works, artistName, nil
}

func getArtistWorksProd(artistId string, first int) ([]IMDBWork, string, error) {
	return getArtistWorks(artistId, first, http.Post)
}

func (IMDB) Parse(options *parser.Options) (*feeds.Feed, error) {
	artistId := options.Get("artistId").(string)
	if artistId == "" {
		return nil, parser.NewNotFoundError("artistId required")
	}
	first := options.Get("first").(int)
	if first <= 0 {
		first = 25
	}

	works, artistName, err := getArtistWorksProd(artistId, first)
	if err != nil {
		return nil, err
	}

	feed := feeds.Feed{
		Title:       fmt.Sprintf("IMDB - %s", artistName),
		Description: fmt.Sprintf("IMDB filmography for %s", artistName),
		Link:        &feeds.Link{Href: fmt.Sprintf("https://www.imdb.com/name/%s/", artistId)},
	}

	for _, w := range works {
		yearStr := ""
		if w.Year > 0 {
			yearStr = fmt.Sprintf(" (%d)", w.Year)
		}
		title := fmt.Sprintf("%s%s", w.Title, yearStr)
		if w.TitleType != "" {
			title = fmt.Sprintf("%s [%s]", title, w.TitleType)
		}

		var descParts []string
		if w.Category != "" {
			descParts = append(descParts, fmt.Sprintf("Credit: %s", w.Category))
		}
		if w.Role != "" {
			descParts = append(descParts, fmt.Sprintf("Role: %s", w.Role))
		}
		if !w.ReleaseDate.IsZero() {
			if w.HasFullDate {
				descParts = append(descParts, fmt.Sprintf("Release: %s", w.ReleaseDate.Format("2006-01-02")))
			} else {
				descParts = append(descParts, fmt.Sprintf("Release year: %d", w.Year))
			}
		}

		created := w.ReleaseDate
		if created.IsZero() {
			created = time.Now()
		}

		item := feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: w.Link},
			Description: strings.Join(descParts, " | "),
			Id:          parser.GetGuid([]string{artistId, w.TitleID, w.Category}),
			Created:     created,
			Updated:     created,
		}
		if w.ImageURL != "" {
			imgExt := parser.GetFileTypeFromUrl(w.ImageURL)
			if !parser.IsImageType(imgExt) {
				imgExt = "jpg"
			}
			item.Enclosure = &feeds.Enclosure{
				Url:    w.ImageURL,
				Type:   "image/" + imgExt,
				Length: "0",
			}
		}
		feed.Items = append(feed.Items, &item)
	}
	return &feed, nil
}

func (IMDB) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "artistId",
				Required: true,
				Type:     "string",
				Help:     "IMDB artist ID (e.g. nm0000138)",
			},
			{
				Flag:     "first",
				Required: false,
				Type:     "int",
				Help:     "max number of credits to return",
				Default:  "25",
			},
		},
		Parser: IMDB{},
	}
}
