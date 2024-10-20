package bugcrowd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (Bugcrowd) String() string {
	return "bugcrowd"
}

func (Bugcrowd) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "disclosures",
				Required: false,
				Type:     "bool",
				Help:     "Show disclosure reports",
				Value:    "",
			},
			{
				Flag:     "accepted",
				Required: false,
				Type:     "bool",
				Help:     "Show accepted reports",
				Default:  "en",
				Value:    "",
			},
			{
				Flag:     "title",
				Required: false,
				Type:     "string",
				Help:     "Feed title",
				Default:  "Bugcrowd",
				Value:    "",
			},
			{
				Flag:     "description",
				Required: false,
				Type:     "string",
				Help:     "Feed description",
				Default:  "Bugcrowd Crowdstream",
				Value:    "",
			},
		},
		Parser: Bugcrowd{},
	}
}

type bugcrowdItem struct {
	LogoUrl                 string `json:"logo"`
	ProgramName             string `json:"program_name"`
	ProgramCode             string `json:"program_code"`
	InProgress              bool   `json:"in_progress"`
	ProgramPath             string `json:"program_path"`
	VisibilityPublic        bool   `json:"visibility_public"`
	ResearcherUsername      string `json:"researcher_username"`
	ResearcherProfilePath   string `json:"researcher_profile_path"`
	Id                      string `json:"id"`
	Disclosed               bool   `json:"disclosed"`
	Priority                int    `json:"priority"`
	CreatedAt               string `json:"created_at"`
	AcceptedAt              string `json:"accepted_at"`
	ClaimedAt               string `json:"claimed_at"`
	ClosedAt                string `json:"closed_at"`
	DisclosedAt             string `json:"disclosed_at"`
	Target                  string `json:"target"`
	SubmissionStateText     string `json:"submission_state_text"`
	SubmissionStateDateText string `json:"submission_state_date_text"`
	Points                  int    `json:"points"`
	Amount                  string `json:"amount"`
	IsForEngagement         bool   `json:"isForEngagement"`
	DisclosureReportUrl     string `json:"disclosure_report_url"`
	Title                   string `json:"title"`
	Substate                string `json:"substate"`
}

type bugcrowdFeed struct {
	Results         []bugcrowdItem `json:"results"`
	CutoffDateLabel string         `json:"cutoff_date_label"`
	PaginationMeta  struct {
		TotalPages int `json:"total_pages"`
	} `json:"pagination_meta"`
}

func (i *bugcrowdItem) GetUpdatedAt() time.Time {
	dates := []time.Time{}
	if i.AcceptedAt != "" {
		d, err := time.Parse("02 Jan 2006", i.AcceptedAt)
		if err == nil {
			dates = append(dates, d)
		}
	}
	if i.ClaimedAt != "" {
		d, err := time.Parse(time.RFC3339, i.ClaimedAt)
		if err == nil {
			dates = append(dates, d)
		}
	}
	if i.ClosedAt != "" {
		d, err := time.Parse(time.RFC3339, i.ClosedAt)
		if err == nil {
			dates = append(dates, d)
		}
	}
	if i.DisclosedAt != "" {
		d, err := time.Parse("02 Jan 2006", i.DisclosedAt)
		if err == nil {
			dates = append(dates, d)
		}
	}
	if i.CreatedAt != "" {
		d, err := time.Parse(time.RFC3339, i.CreatedAt)
		if err == nil {
			dates = append(dates, d)
		}
	}
	if len(dates) == 0 {
		return time.Now()
	}
	return parser.GetLatestDate(dates)
}

func buildReportUrl(item *bugcrowdItem) *feeds.Link {
	if item.DisclosureReportUrl != "" {
		return &feeds.Link{Href: fmt.Sprintf("https://bugcrowd.com%s", item.DisclosureReportUrl)}
	}
	return &feeds.Link{Href: fmt.Sprintf("https://bugcrowd.com%s", item.ProgramPath)}
}

func buildItemTitle(item *bugcrowdItem) string {
	title := item.ProgramName
	if item.CreatedAt != "" {
		t, err := time.Parse(time.RFC3339, item.CreatedAt)
		if err == nil {
			title = fmt.Sprintf("[%s] %s", t.Format("2006-01-02"), title)
		}
	}
	if item.Priority > 0 {
		title = fmt.Sprintf("%s | P%d", title, item.Priority)
	}
	if item.Amount != "" {
		title = fmt.Sprintf("%s | %s", title, item.Amount)
	}
	if item.Title != "" {
		title = fmt.Sprintf("%s | %s", title, item.Title)
	}
	if !item.Disclosed && item.SubmissionStateText != "" {
		title = fmt.Sprintf("%s | %s", title, item.SubmissionStateText)
	} else if item.Disclosed {
		if item.SubmissionStateDateText != "" {
			title = fmt.Sprintf("%s | %s", title, item.SubmissionStateDateText)
		} else {
			title = fmt.Sprintf("%s | Disclosed", title)
		}
	}
	return title
}

func getRewardString(item *bugcrowdItem) string {
	if item.Amount != "" {
		if item.Points > 0 {
			return fmt.Sprintf("%s | %d points", item.Amount, item.Points)
		}
		return item.Amount
	} else if item.Points > 0 {
		return fmt.Sprintf("%d points", item.Points)
	}
	return ""
}

func buildItemContent(item *bugcrowdItem) string {
	description := fmt.Sprintf("Program: %s\n", item.ProgramName)
	reward := getRewardString(item)
	if item.Target != "" {
		description = fmt.Sprintf("%sTarget: %s\n", description, item.Target)
	}
	if item.ResearcherUsername != "" {
		description = fmt.Sprintf("%sReporter: %s\n", description, item.ResearcherUsername)
	}
	if reward != "" {
		description = fmt.Sprintf("%sReward: %s\n", description, reward)
	}
	if !item.Disclosed && item.SubmissionStateText != "" {
		description = fmt.Sprintf("%sState: %s\n", description, item.SubmissionStateText)
	}
	if item.SubmissionStateDateText != "" {
		description = fmt.Sprintf("%sState Date: %s\n", description, item.SubmissionStateDateText)
	}
	if item.DisclosedAt != "" {
		description = fmt.Sprintf("%sDisclosed At: %s\n", description, item.DisclosedAt)
	}
	if item.ClaimedAt != "" {
		description = fmt.Sprintf("%sClaimed At: %s\n", description, item.ClaimedAt)
	}
	if item.AcceptedAt != "" {
		description = fmt.Sprintf("%sAccepted At: %s\n", description, item.AcceptedAt)
	}
	if item.ClosedAt != "" {
		description = fmt.Sprintf("%sClosed At: %s\n", description, item.ClosedAt)
	}
	if item.DisclosureReportUrl != "" {
		description = fmt.Sprintf("%sReport: %s\n", description, item.DisclosureReportUrl)
	}
	return description
}

func feedAdapter(b *bugcrowdFeed, options *parser.Options) (*feeds.Feed, error) {
	feed := feeds.Feed{
		Title:       options.Get("title").(string),
		Description: options.Get("description").(string),
		Items:       []*feeds.Item{},
		Author:      &feeds.Author{Name: "Bugcrowd"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://bugcrowd.com/crowdstream"},
	}

	for _, item := range b.Results {
		updatedAt := item.GetUpdatedAt()
		newItem := feeds.Item{
			Title:       buildItemTitle(&item),
			Content:     strings.Replace(buildItemContent(&item), "\n", "<br/>", -1),
			Description: buildItemContent(&item),
			Link:        buildReportUrl(&item),
			Created:     updatedAt,
			Id:          fmt.Sprint(updatedAt.Format(time.RFC3339), item.Id),
			Updated:     updatedAt,
		}
		feed.Items = append(feed.Items, &newItem)
	}

	return &feed, nil
}

type Bugcrowd struct{}

func BugcrowdParser() parser.Parser {
	return Bugcrowd{}
}

func getCrowdStreamUrl(options *parser.Options) string {
	filters := []string{}
	disclosures, _ := options.Get("disclosures").(*bool)
	accepted, _ := options.Get("accepted").(*bool)
	if *accepted {
		filters = append(filters, "accepted")
	}
	if *disclosures {
		filters = append(filters, "disclosures")
	}
	if !*disclosures && !*accepted {
		filters = []string{"accepted", "disclosures"}
	}
	return fmt.Sprintf("https://bugcrowd.com/crowdstream.json?page=1&filter_by=%s", strings.Join(filters, "%2C"))
}

func (Bugcrowd) Parse(options *parser.Options) (*feeds.Feed, error) {
	url := getCrowdStreamUrl(options)

	resp, err := parser.HttpGet(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var feed bugcrowdFeed

	if err := json.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	return feedAdapter(&feed, options)
}
