package hackerone

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (Hackerone) String() string {
	return "hackerone"
}

func (Hackerone) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			&parser.Option{
				Flag:     "disclosed_only",
				Required: false,
				Type:     "bool",
				Help:     "Show only disclosed reports",
				Default:  "true",
				Value:    "",
			},
			&parser.Option{
				Flag:     "reports_count",
				Required: false,
				Type:     "int",
				Help:     "Number of reports to display",
				Default:  "50",
				Value:    "",
			},
			&parser.Option{
				Flag:     "title",
				Required: false,
				Type:     "string",
				Help:     "Feed title",
				Default:  "HackerOne",
				Value:    "",
			},
			&parser.Option{
				Flag:     "description",
				Required: false,
				Type:     "string",
				Help:     "Feed description",
				Default:  "Hackerone Hacktivity",
				Value:    "",
			},
		},
		Parser: Hackerone{},
	}
}

type hackeroneItem struct {
	Id         string `json:"id"`
	DatabaseId string `json:"databaseId"`
	TypeName   string `json:"__typename"`
	Reporter   struct {
		Id       string `json:"id"`
		Username string `json:"username"`
		TypeName string `json:"__typename"`
	} `json:"reporter"`
	Votes struct {
		TotalCount int    `json:"total_count"`
		TypeName   string `json:"__typename"`
	} `json:"votes"`
	Upvoted bool `json:"upvoted"`
	Team    struct {
		Id                   string `json:"id"`
		Name                 string `json:"name"`
		Handle               string `json:"handle"`
		TypeName             string `json:"__typename"`
		Url                  string `json:"url"`
		MediumProfilePicture string `json:"medium_profile_picture"`
	} `json:"team"`
	Report struct {
		Id                     string `json:"id"`
		TypeName               string `json:"__typename"`
		DatabaseId             string `json:"databaseId"`
		Title                  string `json:"title"`
		Substate               string `json:"substate"`
		Url                    string `json:"url"`
		CreatedAt              string `json:"created_at"`
		ReportGeneratedContent struct {
			Id                string `json:"id"`
			TypeName          string `json:"__typename"`
			HacktivitySummary string `json:"hacktivity_summary"`
		} `json:"report_generated_content"`
	} `json:"report"`
	LatestDisclosableAction     string  `json:"latest_disclosable_action"`
	LatestDisclosableActivityAt string  `json:"latest_disclosable_activity_at"`
	TotalAwardedAmount          float64 `json:"total_awarded_amount"`
	SeverityRating              string  `json:"severity_rating"`
	Currency                    string  `json:"currency"`
}

type hackeroneFeed struct {
	Data struct {
		HacktivityItems struct {
			Edges []struct {
				Node hackeroneItem `json:"node"`
			} `json:"edges"`
		} `json:"hacktivity_items"`
	} `json:"data"`
}

var hackeroneSeverity = map[string]string{
	"critical": "P1",
	"high":     "P2",
	"medium":   "P3",
	"low":      "P4",
	"none":     "P5",
}

var hackeroneCurrency = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
}

func buildItemTitle(item *hackeroneItem) string {
	title := item.Team.Name
	if item.Report.CreatedAt != "" {
		t, err := time.Parse(time.RFC3339, item.Report.CreatedAt)
		if err == nil {
			title = fmt.Sprintf("[%s] %s", t.Format("2006-01-02"), title)
		}
	}
	if item.SeverityRating != "" {
		title = fmt.Sprintf("%s | %s", title, hackeroneSeverity[item.SeverityRating])
	}
	if item.TotalAwardedAmount != 0 {
		title = fmt.Sprintf("%s | %s%d", title, hackeroneCurrency[item.Currency], int(item.TotalAwardedAmount))
	}
	if item.Report.Title != "" {
		title = fmt.Sprintf("%s | %s", title, item.Report.Title)
	}
	if item.Report.Substate != "" {
		title = fmt.Sprintf("%s | %s", title, item.Report.Substate)
	}
	return title
}

func buildItemContent(item *hackeroneItem) string {
	description := fmt.Sprintf("Program: %s\n", item.Team.Name)
	if item.Reporter.Username != "" {
		description = fmt.Sprintf("%sReporter: %s\n", description, item.Reporter.Username)
	}
	if item.TotalAwardedAmount != 0 {
		description = fmt.Sprintf("%sReward: %s%d\n", description, hackeroneCurrency[item.Currency], int(item.TotalAwardedAmount))
	}
	if item.LatestDisclosableActivityAt != "" && item.LatestDisclosableAction != "" {
		description = fmt.Sprintf("%s%s on %s\n", description, item.LatestDisclosableAction, item.LatestDisclosableActivityAt)
	}
	if item.Report.ReportGeneratedContent.HacktivitySummary != "" {
		description = fmt.Sprintf("%sReport: %s\n", description, item.Report.ReportGeneratedContent.HacktivitySummary)
	}
	return description
}

func feedAdapter(b *hackeroneFeed, options *parser.Options) (*feeds.Feed, error) {
	feed := feeds.Feed{
		Title:       options.Get("title").(string),
		Description: options.Get("description").(string),
		Items:       []*feeds.Item{},
		Author:      &feeds.Author{Name: "hackerone"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://hackerone.com/hacktivity/overview"},
	}

	for _, edge := range b.Data.HacktivityItems.Edges {
		item := edge.Node
		updatedAt, err := time.Parse(time.RFC3339, item.LatestDisclosableActivityAt)
		if err != nil {
			fmt.Println("error parsing date", err, item.LatestDisclosableActivityAt)
			continue
		}
		if item.Report.Url == "" {
			if *options.Get("disclosed_only").(*bool) {
				fmt.Printf("skipping disclosed item without a report url %v\n", item)
				continue
			}
			item.Report.Url = item.Team.Url
		}
		newItem := feeds.Item{
			Title:       buildItemTitle(&item),
			Content:     strings.Replace(buildItemContent(&item), "\n", "<br/>", -1),
			Description: buildItemContent(&item),
			Link:        &feeds.Link{Href: item.Report.Url},
			Created:     updatedAt,
			Id:          fmt.Sprint(updatedAt.Format(time.RFC3339), item.Id),
			Updated:     updatedAt,
		}
		feed.Items = append(feed.Items, &newItem)
	}

	return &feed, nil
}

type Hackerone struct{}

func HackeroneParser() parser.Parser {
	return Hackerone{}
}

func (Hackerone) Parse(options *parser.Options) (*feeds.Feed, error) {
	resp, err := hacktivityFeedQuery(options)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var feed hackeroneFeed

	if err := json.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	return feedAdapter(&feed, options)
}
