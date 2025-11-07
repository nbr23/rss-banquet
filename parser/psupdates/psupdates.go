package psupdates

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"time"

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

func parsePSPortalLatestUpdate(doc *goquery.Document, url string, hardware string) (*feeds.Item, error) {
	var item *feeds.Item

	doc.Find("h3").EachWithBreak(func(i int, h3 *goquery.Selection) bool {
		text := strings.TrimSpace(h3.Text())
		text = strings.ReplaceAll(text, "\u00a0", " ")
		versionRegex := regexp.MustCompile(`Version\s*:?\s*([0-9.]+)`)
		matches := versionRegex.FindStringSubmatch(text)

		if len(matches) == 2 {
			version := matches[1]
			paragraphDiv := h3.Closest(".txt-block__paragraph")

			content := ""
			description := ""

			paragraphDiv.Find("ul, p").Each(func(j int, content_s *goquery.Selection) {
				html, _ := content_s.Html()
				content += html
				description += content_s.Text() + "\n"
			})

			item = &feeds.Item{
				Title:       fmt.Sprintf("%s Update: %s", hardware, version),
				Content:     content,
				Description: strings.TrimSpace(description),
				Link:        &feeds.Link{Href: url},
				Id:          guid([]string{url, version}),
				Created:     time.Now(),
			}

			return false
		}
		return true
	})

	if item == nil {
		return nil, fmt.Errorf("unable to find PS Portal version information")
	}

	return item, nil
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

	hardware := options.Get("hardware").(string)
	hardwareUpper := strings.ToUpper(hardware)
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

	if strings.ToLower(hardware) == "psportal" {
		update, err := parsePSPortalLatestUpdate(doc, url, hardwareUpper)
		if err != nil {
			return nil, err
		}
		feed.Items = append(feed.Items, update)
	} else {
		var update feeds.Item

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

		update.Title = fmt.Sprintf("%s Update: %s", hardwareUpper, versionName)
		update.Content, err = releaseDiv.Html()
		update.Description = releaseDiv.Text()
		if err != nil {
			update.Content = fmt.Sprintf("The %s software update %s was released on %v", hardwareUpper, versionName, update.Created)
		}
		update.Link = &feeds.Link{Href: url}
		update.Id = guid([]string{url, versionName})

		feed.Items = append(feed.Items, &update)
	}

	feed.Title = fmt.Sprintf("%s Updates", hardwareUpper)
	feed.Description = fmt.Sprintf("The latest %s updates", hardwareUpper)
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
