package psupdates

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"regexp"

	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"
)

func parseLatestVersion(doc goquery.Document) (string, error) {
	var latestVersion string
	var err error

	doc.Find("div .accordion div .parbase.textblock div p b").Each(func(i int, s *goquery.Selection) {
		// Assuming the first paragraph with version in the text is the latest version
		matched, err := regexp.MatchString("[Vv]ersion", s.Text())
		if err == nil && matched && latestVersion == "" {
			latestVersion = strings.TrimSpace(s.Text())
		}
	})
	if len(latestVersion) == 0 {
		err = fmt.Errorf("unable to parse the latest version in the page")
	}
	return latestVersion, err
}

func guid(hardware string, releaseDate string, versionName string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(hardware, releaseDate, versionName))))
}

func parsePublishDate(doc goquery.Document) (time.Time, error) {
	var publishTimestamp int64
	var err error

	// Find the document publish date metadata
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		if name == "publish_date_timestamp" {
			pubDate, _ := s.Attr("content")
			publishTimestamp, err = strconv.ParseInt(pubDate, 10, 64)
		}
	})
	return time.Unix(publishTimestamp, 0), err
}

func getHardwareURL(hardware string, local string) string {
	return fmt.Sprintf("https://www.playstation.com/%s/support/hardware/%s/system-software/", strings.ToLower(local), strings.ToLower(hardware))
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

	update.Created, err = parsePublishDate(*doc)
	if err != nil {
		return nil, err
	}

	versionName, err := parseLatestVersion(*doc)
	if err != nil {
		return nil, err
	}
	update.Title = fmt.Sprintf("%s Update: %s", hardware, versionName)
	update.Description = fmt.Sprintf("The %s software update %s was released on %v", hardware, versionName, update.Created)
	update.Link = &feeds.Link{Href: url}
	update.Id = guid(hardware, update.Created.Format(time.RFC3339), versionName)

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
