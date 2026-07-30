package imdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

const graphqlEndpoint = "https://caching.graphql.imdb.com/"

const (
	creditsPageSize   = 250
	maxCreditsFetched = 1000
)

type IMDB struct{}

func (IMDB) String() string {
	return "imdb"
}

func IMDBParser() parser.Parser {
	return IMDB{}
}

type IMDBCredit struct {
	Category   string
	CategoryID string
	Role       string
}

type IMDBWork struct {
	TitleID     string
	Title       string
	Year        int
	ReleaseDate time.Time
	HasFullDate bool
	TitleType   string
	TitleTypeID string
	Link        string
	ImageURL    string
	Credits     []IMDBCredit
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
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"titleType"`
		PrimaryImage *struct {
			URL string `json:"url"`
		} `json:"primaryImage"`
	} `json:"title"`
	Category *struct {
		ID   string `json:"id"`
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
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
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

func buildQuery(artistId, after string, first int) []byte {
	variables := map[string]any{
		"id":    artistId,
		"first": first,
	}
	if after != "" {
		variables["after"] = after
	}
	gql := map[string]any{
		"query": `query NameCredits($id: ID!, $first: Int!, $after: ID) {
			name(id: $id) {
				nameText { text }
				credits(first: $first, after: $after, sort: {by: RELEASE_DATE, order: DESC}) {
					pageInfo { hasNextPage endCursor }
					edges {
						node {
							title {
								id
								titleText { text }
								releaseYear { year }
								releaseDate { day month year }
								titleType { id text }
								primaryImage { url }
							}
							category { id text }
							... on Cast { characters { name } }
						}
					}
				}
			}
		}`,
		"variables": variables,
	}
	b, _ := json.Marshal(gql)
	return b
}

func fetchCreditsPage(artistId, after string, post httpPostFunc) (*gqlResponse, error) {
	body := buildQuery(artistId, after, creditsPageSize)
	resp, err := post(graphqlEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, parser.NewInternalError(fmt.Sprintf("imdb graphql request failed: %s", err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, parser.NewInternalError("unable to read imdb graphql response")
	}

	var data gqlResponse
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, parser.NewInternalError("unable to parse imdb graphql response")
	}
	if len(data.Errors) > 0 {
		return nil, parser.NewInternalError(fmt.Sprintf("imdb graphql error: %s", data.Errors[0].Message))
	}
	return &data, nil
}

func creditFromNode(n gqlCredit) IMDBCredit {
	credit := IMDBCredit{}
	if n.Category != nil {
		credit.Category = n.Category.Text
		credit.CategoryID = n.Category.ID
	}
	if len(n.Characters) > 0 {
		names := make([]string, 0, len(n.Characters))
		for _, c := range n.Characters {
			if c.Name != "" {
				names = append(names, c.Name)
			}
		}
		credit.Role = strings.Join(names, ", ")
	}
	return credit
}

func workFromNode(n gqlCredit) IMDBWork {
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
		work.TitleTypeID = n.Title.TitleType.ID
	}
	if n.Title.PrimaryImage != nil {
		work.ImageURL = n.Title.PrimaryImage.URL
	}
	return work
}

func getArtistWorks(artistId string, post httpPostFunc) ([]IMDBWork, string, error) {
	var works []IMDBWork
	byTitle := make(map[string]int)
	artistName := ""
	after := ""
	fetched := 0

	for {
		data, err := fetchCreditsPage(artistId, after, post)
		if err != nil {
			return nil, "", err
		}

		if artistName == "" {
			artistName = data.Data.Name.NameText.Text
			if artistName == "" {
				return nil, "", parser.NewNotFoundError("artist not found")
			}
		}

		credits := data.Data.Name.Credits
		for _, edge := range credits.Edges {
			n := edge.Node
			if n.Title.ID == "" {
				continue
			}
			fetched++
			credit := creditFromNode(n)
			if idx, ok := byTitle[n.Title.ID]; ok {
				works[idx].Credits = append(works[idx].Credits, credit)
				continue
			}
			work := workFromNode(n)
			work.Credits = []IMDBCredit{credit}
			byTitle[n.Title.ID] = len(works)
			works = append(works, work)
		}

		if !credits.PageInfo.HasNextPage || credits.PageInfo.EndCursor == "" || fetched >= maxCreditsFetched {
			break
		}
		after = credits.PageInfo.EndCursor
	}

	return works, artistName, nil
}

// imdb's caching.graphql.imdb.com endpoint rejects requests without a
// Referer header pointing at imdb.com, returning a 403 HTML page instead of
// a JSON body.
func postWithImdbHeaders(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Referer", "https://www.imdb.com/")
	return http.DefaultClient.Do(req)
}

func getArtistWorksProd(artistId string) ([]IMDBWork, string, error) {
	return getArtistWorks(artistId, postWithImdbHeaders)
}

func toFilterSet(values []string) map[string]bool {
	set := make(map[string]bool)
	for _, v := range values {
		v = strings.ToLower(strings.TrimSpace(v))
		if v != "" {
			set[v] = true
		}
	}
	return set
}

func hasMatchingCategory(credits []IMDBCredit, categorySet map[string]bool) bool {
	for _, c := range credits {
		if categorySet[strings.ToLower(c.CategoryID)] {
			return true
		}
	}
	return false
}

func filterWorks(works []IMDBWork, titleTypes, categories []string) []IMDBWork {
	typeSet := toFilterSet(titleTypes)
	categorySet := toFilterSet(categories)
	if len(typeSet) == 0 && len(categorySet) == 0 {
		return works
	}

	filtered := make([]IMDBWork, 0, len(works))
	for _, w := range works {
		if len(typeSet) > 0 && !typeSet[strings.ToLower(w.TitleTypeID)] {
			continue
		}
		if len(categorySet) > 0 && !hasMatchingCategory(w.Credits, categorySet) {
			continue
		}
		filtered = append(filtered, w)
	}
	return filtered
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

	works, artistName, err := getArtistWorksProd(artistId)
	if err != nil {
		return nil, err
	}

	works = filterWorks(works, options.Get("titleType").([]string), options.Get("category").([]string))
	sort.SliceStable(works, func(i, j int) bool {
		return works[i].ReleaseDate.After(works[j].ReleaseDate)
	})
	if len(works) > first {
		works = works[:first]
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
		var categories, roles []string
		for _, c := range w.Credits {
			if c.Category != "" {
				categories = append(categories, c.Category)
			}
			if c.Role != "" {
				roles = append(roles, c.Role)
			}
		}
		if len(categories) > 0 {
			label := "Credit"
			if len(categories) > 1 {
				label = "Credits"
			}
			descParts = append(descParts, fmt.Sprintf("%s: %s", label, strings.Join(categories, ", ")))
		}
		if len(roles) > 0 {
			descParts = append(descParts, fmt.Sprintf("Role: %s", strings.Join(roles, ", ")))
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
			Id:          parser.GetGuid([]string{artistId, w.TitleID}),
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
				Help:     "max number of titles to return",
				Default:  "25",
			},
			{
				Flag:     "titleType",
				Required: false,
				Type:     "stringSlice",
				Help:     "filter by title type (e.g. movie, short, tvSeries, tvMiniSeries, tvMovie, tvSpecial, tvShort, video, videoGame, musicVideo, podcastSeries)",
			},
			{
				Flag:     "category",
				Required: false,
				Type:     "stringSlice",
				Help:     "filter by credit category (e.g. actor, director, writer, producer, self, soundtrack, archive_footage, art_department, animation_department, miscellaneous, thanks)",
			},
		},
		Parser: IMDB{},
	}
}
