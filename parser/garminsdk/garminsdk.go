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

func (GarminSDK) String() string {
	return "garminsdk"
}

func (GarminSDK) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			&parser.Option{
				Flag:     "sdks",
				Required: false,
				Type:     "stringSlice",
				Help:     "list of names of the sdks to watch: fit, connect-iq",
				Default:  "fit",
				Value:    "",
			},
		},
		Parser: GarminSDK{},
	}
}

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

func getDownloadButtons(s *goquery.Document) []*goquery.Selection {
	buttons := s.Find("a.btn")
	res := make([]*goquery.Selection, 0)
	for i := range buttons.Nodes {
		if strings.Contains(buttons.Eq(i).Text(), "Accept & Download") {
			res = append(res, buttons.Eq(i))
		}
	}
	return res
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

func (GarminSDK) Parse(options *parser.Options) (*feeds.Feed, error) {
	sdkNames := options.Get("sdks").([]string)
	var feed feeds.Feed

	for _, sdkName := range sdkNames {
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
		downloadButtons := getDownloadButtons(doc)
		if len(downloadButtons) == 0 {
			return nil, fmt.Errorf("unable to find the download")
		}

		for _, downloadButton := range downloadButtons {
			var update feeds.Item

			downloadUrl := downloadButton.AttrOr("href", "")
			downloadName := downloadButton.AttrOr("download", "")
			if downloadUrl == "" || downloadName == "" {
				return nil, fmt.Errorf("unable to find the download")
			}

			update.Created, err = time.Parse("January 2, 2006", releaseDate)
			if err != nil {
				return nil, err
			}
			update.Title = fmt.Sprintf("[%s] Garmin %s SDK Update: %s", releaseDate, sdkName, downloadName)
			update.Description = fmt.Sprintf("The Garmin %s SDK update %s was released on %v", sdkName, downloadName, update.Created)
			update.Link = &feeds.Link{Href: downloadUrl}
			update.Id = parser.GetGuid([]string{downloadUrl, releaseDate})
			feed.Items = append(feed.Items, &update)
		}
	}

	feed.Title = fmt.Sprintf("Garmin %s SDK Updates", strings.Join(sdkNames, ", "))
	feed.Description = fmt.Sprintf("The latest Garmin %s SDK updates", strings.Join(sdkNames, ", "))

	feed.Author = &feeds.Author{
		Name: "Garmin",
	}
	feed.Link = &feeds.Link{Href: "https://developer.garmin.com/"}

	return &feed, nil
}

type GarminSDK struct{}

func GarminSDKParser() parser.Parser {
	return GarminSDK{}
}
