package garminwearables

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"
)

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

func (GarminWearables) Parse(options map[string]any) (*feeds.Feed, error) {
	var feed feeds.Feed
	resp, err := http.Get("https://www.garmin.com/en-US/support/software/wearables/")

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return nil, err
	}

	releaseNotes := getReleaseNotes(doc)

	for _, releaseNote := range releaseNotes {
		var update feeds.Item

		update.Created, err = parser.GetRemoteFileLastModified(releaseNote[0])
		if err != nil {
			return nil, err
		}
		update.Title = fmt.Sprintf("[%s] Garmin Wearable Update", releaseNote[2])
		update.Description = fmt.Sprintf("A Garmin Wearable update was released on %v", update.Created)
		update.Link = &feeds.Link{Href: releaseNote[0]}
		update.Id = parser.GetGuid([]string{releaseNote[0], releaseNote[2]})
		feed.Items = append(feed.Items, &update)
	}

	feed.Title = parser.DefaultedGet(options, "title", "Garmin %s Wearable Updates")
	feed.Description = parser.DefaultedGet(options, "description", "The latest Garmin %s Wearable updates")

	feed.Author = &feeds.Author{
		Name: "Garmin",
	}
	feed.Link = &feeds.Link{Href: "https://www.garmin.com/en-US/support/software/wearables/"}

	return &feed, nil
}

func (GarminWearables) Help() string {
	return ""
}

type GarminWearables struct{}

func GarminWearablesParser() parser.Parser {
	return GarminWearables{}
}
