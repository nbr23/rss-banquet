package garminsdk

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"
)

func parseReleaseDate(s *goquery.Document) (string, error) {
	disclaimer := s.Find("p.disclaimer")
	for i := range disclaimer.Nodes {
		if strings.Contains(disclaimer.Eq(i).Text(), "Release Date: ") ||
			strings.Contains(disclaimer.Eq(i).Text(), "Last Updated:") {
			return strings.Split(disclaimer.Eq(i).Text(), ": ")[1], nil
		}
	}
	return "", fmt.Errorf("unable to parse the latest version in the page")
}

func getDownloadButton(s *goquery.Document) *goquery.Selection {
	buttons := s.Find("a.btn")
	for i := range buttons.Nodes {
		if strings.Contains(buttons.Eq(i).Text(), "Accept & Download") {
			return buttons.Eq(i)
		}
	}
	return nil
}

func getValidUrl(sdkName string) (string, *http.Response, error) {
	url := fmt.Sprintf("https://developer.garmin.com/%s/download/", strings.ToLower(sdkName))
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	if resp.StatusCode == 404 {
		url = fmt.Sprintf("https://developer.garmin.com/%s/sdk/", strings.ToLower(sdkName))
		resp, err = http.Get(url)
	}
	if err != nil {
		return "", nil, err
	}
	if resp.StatusCode != 200 {
		return "", nil, fmt.Errorf("unable to fetch the update page, status code: %d", resp.StatusCode)
	}
	return url, resp, nil
}

func (GarminSDK) Parse(options map[string]any) (*feeds.Feed, error) {
	sdkNames := parser.DefaultedGetSlice(options, "sdks", []string{"fit"})
	var feed feeds.Feed

	for _, sdkName := range sdkNames {
		var update feeds.Item

		_, resp, err := getValidUrl(sdkName)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, err
		}

		releaseDate, err := parseReleaseDate(doc)
		if err != nil {
			return nil, err
		}
		downloadButton := getDownloadButton(doc)
		if downloadButton == nil {
			return nil, fmt.Errorf("unable to find the download")
		}
		downloadUrl := downloadButton.AttrOr("href", "")
		downloadName := downloadButton.AttrOr("download", "")
		if downloadUrl == "" || downloadName == "" {
			return nil, fmt.Errorf("unable to find the download")
		}

		update.Created, err = time.Parse("January 2, 2006", releaseDate)
		if err != nil {
			return nil, err
		}
		update.Title = fmt.Sprintf("Garmin %s SDK Update: %s", sdkName, downloadName)
		update.Description = fmt.Sprintf("The Garmin %s SDK update %s was released on %v", sdkName, downloadName, update.Created)
		update.Link = &feeds.Link{Href: downloadUrl}
		update.Id = parser.GetGuid([]string{downloadUrl, releaseDate})
		feed.Items = append(feed.Items, &update)
	}

	feed.Title = parser.DefaultedGet(options, "title", fmt.Sprintf("Garmin %s SDK Updates", strings.Join(sdkNames, ", ")))
	feed.Description = parser.DefaultedGet(options, "description", fmt.Sprintf("The latest Garmin %s SDK updates", strings.Join(sdkNames, ", ")))

	feed.Author = &feeds.Author{
		Name: "Garmin",
	}
	feed.Link = &feeds.Link{Href: "https://developer.garmin.com/"}

	return &feed, nil
}

func (GarminSDK) Help() string {
	return "\toptions:\n" +
		"\t - sdk: name of the sdk to watch (default: fit)\n"
}

type GarminSDK struct{}

func GarminSDKParser() parser.Parser {
	return GarminSDK{}
}
