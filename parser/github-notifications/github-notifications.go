package githubnotifications

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

type GitHubNotifications struct{}

func (GitHubNotifications) String() string {
	return "github-notifications"
}

func (GitHubNotifications) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "all",
				Required: false,
				Type:     "bool",
				Help:     "Include read notifications",
				Default:  "false",
			},
			{
				Flag:     "participating",
				Required: false,
				Type:     "bool",
				Help:     "Only show notifications where user is directly participating or mentioned",
				Default:  "false",
			},
			{
				Flag:     "age",
				Required: false,
				Type:     "string",
				Help:     "Only show notifications from the past duration (e.g., 24h, 7d, 2w)",
				Default:  "",
			},
			{
				Flag:     "before",
				Required: false,
				Type:     "string",
				Help:     "Only show notifications updated before this ISO 8601 timestamp (YYYY-MM-DDTHH:MM:SSZ)",
				Default:  "",
			},
			{
				Flag:     "org",
				Required: false,
				Type:     "string",
				Help:     "Filter by organization/owner name (client-side filter)",
				Default:  "",
			},
			{
				Flag:     "reason",
				Required: false,
				Type:     "stringSlice",
				Help:     "Filter by notification reason: assign, author, comment, ci_activity, invitation, manual, mention, review_requested, security_alert, state_change, subscribed, team_mention (client-side filter)",
				Default:  "",
			},
			{
				Flag:     "title",
				Required: false,
				Type:     "string",
				Help:     "Feed title",
				Default:  "GitHub Notifications",
			},
			{
				Flag:     "description",
				Required: false,
				Type:     "string",
				Help:     "Feed description",
				Default:  "GitHub Notifications Feed",
			},
		},
		Parser: GitHubNotifications{},
	}
}

func GitHubNotificationsParser() parser.Parser {
	return GitHubNotifications{}
}

type notificationOwner struct {
	Login string `json:"login"`
}

type notificationRepository struct {
	FullName string            `json:"full_name"`
	Owner    notificationOwner `json:"owner"`
	HTMLURL  string            `json:"html_url"`
}

type notificationSubject struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type notification struct {
	ID         string                 `json:"id"`
	Repository notificationRepository `json:"repository"`
	Subject    notificationSubject    `json:"subject"`
	Reason     string                 `json:"reason"`
	Unread     bool                   `json:"unread"`
	UpdatedAt  string                 `json:"updated_at"`
	URL        string                 `json:"url"`
}

func parseAge(age string) (time.Time, error) {
	if age == "" {
		return time.Time{}, nil
	}

	age = strings.TrimSpace(strings.ToLower(age))
	if len(age) < 2 {
		return time.Time{}, fmt.Errorf("invalid age format: %s", age)
	}

	unit := age[len(age)-1]
	valueStr := age[:len(age)-1]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid age value: %s", age)
	}

	now := time.Now().UTC()
	switch unit {
	case 'h':
		return now.Add(-time.Duration(value) * time.Hour), nil
	case 'd':
		return now.AddDate(0, 0, -value), nil
	case 'w':
		return now.AddDate(0, 0, -value*7), nil
	default:
		return time.Time{}, fmt.Errorf("invalid age unit: %c (use h, d, or w)", unit)
	}
}

func buildAPIURL(options *parser.Options) string {
	baseURL := "https://api.github.com/notifications"
	params := url.Values{}

	if options.GetBool("all") {
		params.Set("all", "true")
	}

	if options.GetBool("participating") {
		params.Set("participating", "true")
	}

	age := options.Get("age").(string)
	if age != "" {
		since, err := parseAge(age)
		if err == nil && !since.IsZero() {
			params.Set("since", since.Format(time.RFC3339))
		}
	}

	before := options.Get("before").(string)
	if before != "" {
		params.Set("before", before)
	}

	params.Set("per_page", "50")

	if len(params) > 0 {
		return baseURL + "?" + params.Encode()
	}
	return baseURL
}

func convertAPIURLToWebURL(apiURL string) string {
	if apiURL == "" {
		return ""
	}

	re := regexp.MustCompile(`https://api\.github\.com/repos/([^/]+/[^/]+)/(issues|pulls|commits)/(\d+|[a-f0-9]+)`)
	matches := re.FindStringSubmatch(apiURL)
	if len(matches) == 4 {
		repo := matches[1]
		resourceType := matches[2]
		id := matches[3]
		if resourceType == "pulls" {
			resourceType = "pull"
		} else if resourceType == "commits" {
			resourceType = "commit"
		}
		return fmt.Sprintf("https://github.com/%s/%s/%s", repo, resourceType, id)
	}

	return strings.Replace(apiURL, "https://api.github.com/repos/", "https://github.com/", 1)
}

func formatReason(reason string) string {
	replacer := strings.NewReplacer("_", " ")
	return replacer.Replace(reason)
}

func buildItemTitle(n *notification) string {
	return fmt.Sprintf("[%s] %s", n.Repository.FullName, n.Subject.Title)
}

func buildItemContent(n *notification) string {
	content := fmt.Sprintf("Repository: %s\n", n.Repository.FullName)
	content += fmt.Sprintf("Type: %s\n", n.Subject.Type)
	content += fmt.Sprintf("Reason: %s\n", formatReason(n.Reason))
	if n.Unread {
		content += "Status: Unread\n"
	} else {
		content += "Status: Read\n"
	}
	return content
}

func filterNotifications(notifications []notification, options *parser.Options) []notification {
	org := options.Get("org").(string)
	reasons := options.Get("reason").([]string)

	// Filter out empty strings from reasons
	filteredReasons := make([]string, 0)
	for _, r := range reasons {
		r = strings.TrimSpace(r)
		if r != "" {
			filteredReasons = append(filteredReasons, r)
		}
	}

	if org == "" && len(filteredReasons) == 0 {
		return notifications
	}

	filtered := make([]notification, 0)
	reasonMap := make(map[string]bool)
	for _, r := range filteredReasons {
		reasonMap[r] = true
	}

	for _, n := range notifications {
		if org != "" && !strings.EqualFold(n.Repository.Owner.Login, org) {
			continue
		}

		if len(filteredReasons) > 0 && !reasonMap[n.Reason] {
			continue
		}

		filtered = append(filtered, n)
	}

	return filtered
}

func feedAdapter(notifications []notification, options *parser.Options) (*feeds.Feed, error) {
	feed := feeds.Feed{
		Title:       options.Get("title").(string),
		Description: options.Get("description").(string),
		Items:       []*feeds.Item{},
		Author:      &feeds.Author{Name: "GitHub"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://github.com/notifications"},
	}

	for _, n := range notifications {
		updatedAt, err := time.Parse(time.RFC3339, n.UpdatedAt)
		if err != nil {
			updatedAt = time.Now()
		}

		webURL := convertAPIURLToWebURL(n.Subject.URL)
		if webURL == "" {
			webURL = n.Repository.HTMLURL
		}

		item := feeds.Item{
			Title:       buildItemTitle(&n),
			Content:     strings.Replace(buildItemContent(&n), "\n", "<br/>", -1),
			Description: buildItemContent(&n),
			Link:        &feeds.Link{Href: webURL},
			Created:     updatedAt,
			Updated:     updatedAt,
			Id:          parser.GetGuid([]string{n.ID}),
		}
		feed.Items = append(feed.Items, &item)
	}

	parser.SortFeedEntries(&feed)
	return &feed, nil
}

func (GitHubNotifications) Parse(options *parser.Options) (*feeds.Feed, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, parser.NewInternalError("GitHub token not configured. Set GITHUB_TOKEN environment variable.")
	}

	apiURL := buildAPIURL(options)

	resp, err := parser.HttpGet(apiURL, map[string]any{
		"headers": map[string]string{
			"Authorization":        "Bearer " + token,
			"Accept":               "application/vnd.github+json",
			"X-GitHub-Api-Version": "2022-11-28",
		},
	})

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, parser.NewInternalError("GitHub authentication failed. Check your token.")
	}

	if resp.StatusCode == 403 {
		return nil, parser.NewInternalError("GitHub API rate limit exceeded or insufficient permissions.")
	}

	if resp.StatusCode != 200 {
		return nil, parser.NewInternalError(fmt.Sprintf("GitHub API error: %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var notifications []notification
	if err := json.Unmarshal(data, &notifications); err != nil {
		return nil, err
	}

	notifications = filterNotifications(notifications, options)

	return feedAdapter(notifications, options)
}
