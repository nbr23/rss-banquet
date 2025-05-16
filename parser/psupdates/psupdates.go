package psupdates

import (
	"crypto/sha256"
	"fmt"
	"regexp"

	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (PSUpdates) String() string {
	return "psupdates"
}

func (PSUpdates) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			{
				Flag:     "hardware",
				Required: true,
				Type:     "string",
				Help:     "hardware of the updates",
				Default:  "ps5",
			},
			{
				Flag:     "local",
				Required: false,
				Type:     "string",
				Help:     "local of the updates",
				Default:  "en-us",
			},
		},
		Parser: PSUpdates{},
	}
}

func parseLatestVersion(s *goquery.Selection) (string, error) {
	var latestVersion string
	var err error

	r := regexp.MustCompile(`[Vv]ersion[ ]*:?[ ]*([^ ]+)`)
	matches := r.FindStringSubmatch(s.Text())

	if (matches == nil) || (len(matches) != 2) {
		err = fmt.Errorf("unable to parse the latest version in the page")
	} else {
		latestVersion = strings.TrimSpace(matches[1])
	}
	return latestVersion, err
}

func guid(ss []string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(ss))))
}

func getReleaseDiv(doc goquery.Document) *goquery.Selection {
	return doc.Find(".body-text-block .txt-block-paragraph").First()
}

func getHardwareURL(hardware string, local string) string {
	return fmt.Sprintf("https://www.playstation.com/%s/support/hardware/%s/system-software-info/", strings.ToLower(local), strings.ToLower(hardware))
}

func getUpdateFileUrl(hardware string, local string) (string, error) {
	url := fmt.Sprintf("https://www.playstation.com/%s/support/hardware/%s/system-software/", strings.ToLower(local), strings.ToLower(hardware))
	resp, err := parser.HttpGet(url, nil)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unable to fetch the update page, status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	link := doc.Find("[href*='update.playstation.net']").First()

	href, exists := link.Attr("href")
	if !exists {
		return "", fmt.Errorf("unable to find the update file url")
	}

	return href, nil
}

func (PSUpdates) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed
	var update feeds.Item

	hardware := options.Get("hardware").(string)
	hardware = strings.ToUpper(hardware)
	local := options.Get("local").(string)
	url := getHardwareURL(hardware, local)

	resp, err := parser.HttpGet(url, nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to fetch the update page, status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	releaseDiv := getReleaseDiv(*doc)

	versionName, err := parseLatestVersion(releaseDiv)
	if err != nil {
		return nil, err
	}

	fileUrl, err := getUpdateFileUrl(hardware, local)
	if err != nil {
		return nil, err
	}

	update.Created, err = parser.GetRemoteFileLastModified(fileUrl)
	if err != nil {
		return nil, err
	}

	update.Title = fmt.Sprintf("%s Update: %s", hardware, versionName)
	update.Content, err = releaseDiv.Html()
	update.Description = releaseDiv.Text()
	if err != nil {
		update.Content = fmt.Sprintf("The %s software update %s was released on %v", hardware, versionName, update.Created)
	}
	update.Link = &feeds.Link{Href: url}
	update.Id = guid([]string{url, versionName})

	feed.Title = fmt.Sprintf("%s Updates", hardware)
	feed.Description = fmt.Sprintf("The latest %s updates", hardware)
	feed.Items = append(feed.Items, &update)
	feed.Author = &feeds.Author{
		Name: "PlayStation",
	}
	feed.Link = &feeds.Link{Href: url}

	return &feed, nil
}

type PSUpdates struct{}

func PSUpdatesParser() parser.Parser {
	return PSUpdates{}
}
