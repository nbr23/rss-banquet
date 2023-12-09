package psupdates

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"regexp"

	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"
)

func parseLatestVersion(s *goquery.Selection) (string, error) {
	var latestVersion string
	var err error

	r := regexp.MustCompile(`[Vv]ersion:*(.*)\n`)
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
	resp, err := http.Get(url)
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

func getRemoteFileLastModified(url string) (time.Time, error) {
	resp, err := http.Head(url)
	if err != nil {
		return time.Time{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return time.Time{}, fmt.Errorf("unable to fetch the update file, status code: %d", resp.StatusCode)
	}

	lastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		return time.Time{}, err
	}

	return lastModified, nil
}

func (PSUpdates) Parse(options map[string]any) (*feeds.Feed, error) {
	var feed feeds.Feed
	var update feeds.Item

	hardware := parser.DefaultedGet(options, "hardware", "ps5")
	hardware = strings.ToUpper(hardware)
	local := parser.DefaultedGet(options, "local", "en-us")
	url := getHardwareURL(hardware, local)

	resp, err := http.Get(url)

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

	update.Updated, err = getRemoteFileLastModified(fileUrl)
	if err != nil {
		return nil, err
	}
	update.Created = update.Updated

	update.Title = fmt.Sprintf("%s Update: %s", hardware, versionName)
	update.Description, err = releaseDiv.Html()
	if err != nil {
		update.Description = fmt.Sprintf("The %s software update %s was released on %v", hardware, versionName, update.Created)
	}
	update.Link = &feeds.Link{Href: url}
	update.Id = guid([]string{url, versionName})

	feed.Title = parser.DefaultedGet(options, "title", fmt.Sprintf("%s Updates", hardware))
	feed.Description = parser.DefaultedGet(options, "description", fmt.Sprintf("The latest %s updates", hardware))
	feed.Items = append(feed.Items, &update)
	feed.Author = &feeds.Author{
		Name: "PlayStation",
	}
	feed.Link = &feeds.Link{Href: url}
	feed.Created = update.Created

	return &feed, nil
}

func (PSUpdates) Help() string {
	return "\toptions:\n" +
		"\t - hardware: ps5 or ps4 (default: ps5)\n" +
		"\t - local: en-us or fr-fr (default: en-us)\n"
}

type PSUpdates struct{}

func PSUpdatesParser() parser.Parser {
	return PSUpdates{}
}
