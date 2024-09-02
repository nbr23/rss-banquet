package garminwearables

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (GarminWearables) String() string {
	return "garminwearables"
}

func (GarminWearables) GetOptions() parser.Options {
	return parser.Options{}
}

func getReleaseNotes(s *goquery.Document) [][]string {
	scriptTags := s.Find("script")
	reLink := regexp.MustCompile(`https://www\w?.garmin.com/wearables/PDF/WearablesSoftwareUpdate/([0-9]+)/(\w+).pdf`)
	for i := range scriptTags.Nodes {
		if strings.Contains(scriptTags.Eq(i).Text(), "\"pageContent\"") {
			return reLink.FindAllStringSubmatch(scriptTags.Eq(i).Text(), -1)
		}
	}
	return nil
}

func BruteForcePossibleVersions() []*feeds.Item {
	now := time.Now()
	year := now.Year()

	items := []*feeds.Item{}

	urlFormat := "https://www8.garmin.com/wearables/PDF/WearablesSoftwareUpdate/%d/%s%d.pdf"

	for y := year - 1; y <= year; y++ {
		for m := 1; m <= 12; m++ {
			var update feeds.Item
			var err error
			update.Link = &feeds.Link{Href: fmt.Sprintf(urlFormat, y, time.Month(m).String(), y)}
			update.Created, err = parser.GetRemoteFileLastModified(update.Link.Href)
			if err != nil {
				continue
			}
			update.Title = fmt.Sprintf("[%s%d] Garmin Wearable Update", time.Month(m).String(), y)
			update.Content = fmt.Sprintf("A Garmin Wearable update was released on %v", update.Created)
			update.Description = update.Content
			update.Id = parser.GetGuid([]string{update.Link.Href, fmt.Sprintf("%s%d", time.Month(m).String(), y)})
			items = append(items, &update)
		}
	}
	return items
}

func GetLatestVersions() ([]*feeds.Item, error) {
	resp, err := parser.HttpGet("https://www.garmin.com/en-US/support/software/wearables/")
	items := []*feeds.Item{}

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to fetch the Garmin Wearable updates page, status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return nil, err
	}

	releaseNotes := getReleaseNotes(doc)

	for _, releaseNote := range releaseNotes {
		var update feeds.Item

		releaseDate, err := time.Parse("January2006", releaseNote[2])
		if err != nil {
			return nil, err
		}

		update.Created, err = parser.GetRemoteFileLastModified(releaseNote[0])
		if err != nil {
			return nil, err
		}

		if releaseDate.Month() != update.Created.Month() {
			update.Created = releaseDate
		}

		update.Title = fmt.Sprintf("[%s] Garmin Wearable Update", releaseNote[2])
		update.Content = fmt.Sprintf("A Garmin Wearable update was released on %v", update.Created)
		update.Description = update.Content
		update.Link = &feeds.Link{Href: releaseNote[0]}
		update.Id = parser.GetGuid([]string{releaseNote[0], releaseNote[2]})
		items = append(items, &update)
	}
	return items, nil
}

func (GarminWearables) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed

	var items []*feeds.Item

	items, err := GetLatestVersions()

	if err != nil {
		fmt.Println("Falling back to brute force")
		items = BruteForcePossibleVersions()
	}

	feed.Title = "Garmin Wearable Updates"
	feed.Description = "The latest Garmin Wearable updates"
	feed.Items = items

	feed.Author = &feeds.Author{
		Name: "Garmin",
	}
	feed.Link = &feeds.Link{Href: "https://www.garmin.com/en-US/support/software/wearables/"}

	return &feed, nil
}

type GarminWearables struct{}

func GarminWearablesParser() parser.Parser {
	return GarminWearables{}
}
