package hackeronePrograms

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (HackeronePrograms) String() string {
	return "hackeroneprograms"
}

func (HackeronePrograms) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			&parser.Option{
				Flag:     "results_count",
				Required: false,
				Type:     "int",
				Help:     "Number of programs to display",
				Default:  "50",
				Value:    "",
			},
			&parser.Option{
				Flag:     "title",
				Required: false,
				Type:     "string",
				Help:     "Feed title",
				Default:  "HackerOne Programs",
				Value:    "",
			},
			&parser.Option{
				Flag:     "description",
				Required: false,
				Type:     "string",
				Help:     "Feed description",
				Default:  "Hackerone Program Launch",
				Value:    "",
			},
		},
		Parser: HackeronePrograms{},
	}
}

type hackeroneProgramItem struct {
	Id                      string         `json:"id"`
	Handle                  string         `json:"handle"`
	Name                    string         `json:"name"`
	TeamId                  int            `json:"team_id"`
	TriageActive            bool           `json:"triage_active"`
	AllowsBountySplitting   bool           `json:"allows_bounty_splitting"`
	LaunchedAt              string         `json:"launched_at"`
	State                   string         `json:"state"`
	OffersBounties          bool           `json:"offers_bounties"`
	LastUpdatedAt           string         `json:"last_updated_at"`
	Currency                string         `json:"currency"`
	TeamType                string         `json:"team_type"`
	MinimumBountyTableValue float64        `json:"minimum_bounty_table_value"`
	MaximumBountyTableValue float64        `json:"maximum_bounty_table_value"`
	FirstResponseTime       float64        `json:"first_response_time"`
	SubmissionState         string         `json:"submission_state"`
	ResolvedReportCount     int            `json:"resolved_report_count"`
	GoldStandard            bool           `json:"gold_standard"`
	AwardedReportCount      int            `json:"awarded_report_count"`
	AwardedReporterCount    int            `json:"awarded_reporter_count"`
	StructuredScopeStats    map[string]int `json:"structured_scope_stats"`
	Campaign                struct {
		Id             string  `json:"id"`
		CampaignType   string  `json:"campaign_type"`
		StartDate      string  `json:"start_date"`
		EndDate        string  `json:"end_date"`
		Critical       float64 `json:"critical"`
		TargetAudience bool    `json:"target_audience"`
	} `json:"campaign"`
	H1Clear bool `json:"h1_clear"`
	Idv     bool `json:"idv"`
}

type hackeroneProgramFeed struct {
	Data struct {
		OpportunitiesSearch struct {
			Nodes []hackeroneProgramItem `json:"nodes"`
		} `json:"opportunities_search"`
	} `json:"data"`
}

func getProgramType(item *hackeroneProgramItem) string {
	programType := strings.Split(item.TeamType, "::")
	if len(programType) > 1 {
		switch programType[1] {
		case "BugBountyProgram":
			return "BBP"
		case "VulnerabilityDisclosureProgram":
			return "VDP"
		}
		return programType[1]
	}
	return item.TeamType
}

func buildItemTitle(item *hackeroneProgramItem) string {
	return fmt.Sprintf("[%s] %s launched a program on %s", getProgramType(item), item.Name, item.LaunchedAt)
}

func buildItemContent(item *hackeroneProgramItem) string {
	description := fmt.Sprintf(`
%s launched a program on %s\n
Program type: %s\n
State: %s\n
Resolved reports so far: %d\n
Submission state: %s\n
Program last updated on %s\n
`, item.Name, item.LaunchedAt, item.TeamType, item.State, item.ResolvedReportCount, item.SubmissionState, item.LastUpdatedAt)

	if item.OffersBounties && item.MinimumBountyTableValue != 0 && item.MaximumBountyTableValue != 0 {
		description = fmt.Sprintf("%sBounty range: %d-%d %s", description, int(item.MinimumBountyTableValue), int(item.MaximumBountyTableValue), item.Currency)
	}

	if item.StructuredScopeStats != nil {
		description = fmt.Sprintf("%s\nScope:\n", description)
		for scope := range item.StructuredScopeStats {
			description = fmt.Sprintf("%s- %s: %d\n", description, scope, item.StructuredScopeStats[scope])
		}
	}
	return description
}

func feedAdapter(b *hackeroneProgramFeed, options *parser.Options) (*feeds.Feed, error) {
	feed := feeds.Feed{
		Title:       options.Get("title").(string),
		Description: options.Get("description").(string),
		Items:       []*feeds.Item{},
		Author:      &feeds.Author{Name: "hackerone"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://hackerone.com/opportunities/all/"},
	}

	for _, item := range b.Data.OpportunitiesSearch.Nodes {
		updatedAt, err := time.Parse(time.RFC3339, item.LaunchedAt)
		if err != nil {
			fmt.Println("error parsing date", err, item.LaunchedAt)
			continue
		}
		newItem := feeds.Item{
			Title:       buildItemTitle(&item),
			Content:     strings.Replace(buildItemContent(&item), "\n", "<br/>", -1),
			Description: buildItemContent(&item),
			Link:        &feeds.Link{Href: fmt.Sprintf("https://hackerone.com/%s?type=team", item.Handle)},
			Created:     updatedAt,
			Id:          fmt.Sprint(updatedAt.Format(time.RFC3339), item.Id),
			Updated:     updatedAt,
		}
		feed.Items = append(feed.Items, &newItem)
	}

	return &feed, nil
}

type HackeronePrograms struct{}

func HackeroneProgramsParser() parser.Parser {
	return HackeronePrograms{}
}

func (HackeronePrograms) Parse(options *parser.Options) (*feeds.Feed, error) {
	resp, err := programsFeedQuery(options)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var feed hackeroneProgramFeed

	if err := json.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	return feedAdapter(&feed, options)
}
