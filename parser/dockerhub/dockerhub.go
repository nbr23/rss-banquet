package dockerhub

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"strings"

	"encoding/json"

	"github.com/gorilla/feeds"
	"github.com/nbr23/rss-banquet/parser"
)

func (DockerHub) String() string {
	return "dockerhub"
}

func (DockerHub) GetOptions() parser.Options {
	return parser.Options{
		OptionsList: []*parser.Option{
			&parser.Option{
				Flag:     "image",
				Required: true,
				Type:     "string",
				Help:     "image name (eg nbr23/rss-banquet:latest)",
				Value:    "",
				IsPath:   true,
			},
			&parser.Option{
				Flag:     "platform",
				Required: false,
				Type:     "string",
				Help:     "image platform filter (linux/arm64, ...)",
				Value:    "",
			},
		},
		Parser: DockerHub{},
	}
}

type dockerImageName struct {
	Org   string
	Image string
	Tag   string
}

type dockerImagePlatform struct {
	Os           string
	Architecture string
	Variant      string
}

type dockerhubTag struct {
	Name          string           `json:"name"`
	LastUpdated   string           `json:"last_updated"`
	TagLastPushed string           `json:"tag_last_pushed"`
	Digest        string           `json:"digest"`
	Images        []dockerhubImage `json:"images"`
}

type dockerhubResponse struct {
	Count int `json:"count"`
	// Next     string         `json:"next"`
	// Previous string         `json:"previous"`
	Results []dockerhubTag `json:"results"`
}

type dockerhubImage struct {
	Digest       string `json:"digest"`
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
	Variant      string `json:"variant"`
	LastPushed   string `json:"last_pushed"`
	FullName     dockerImageName
}

func (d dockerImageName) GetImageURL(digest string) string {
	return fmt.Sprintf("https://hub.docker.com/layers/%s/%s/%s/images/%s", d.Org, d.Image, d.Tag, digest)
}

func (d dockerImageName) GetURL() string {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags", d.Org, d.Image)
	if d.Tag != "" {
		url = fmt.Sprintf("%s/%s", url, d.Tag)
	}
	return url
}

func (d dockerImageName) GetUiURL() string {
	return fmt.Sprintf("https://hub.docker.com/r/%s/%s/tags?name=%s", d.Org, d.Image, d.Tag)
}

func (d dockerImageName) String() string {
	return fmt.Sprintf("%s/%s", d.Org, d.Image)
}

func (d dockerImageName) Pretty() string {
	image := d.Image
	if d.Tag != "" {
		image = fmt.Sprintf("%s:%s", image, d.Tag)
	}
	if d.Org == "library" {
		return image
	}
	return fmt.Sprintf("%s/%s", d.Org, image)
}

func (d dockerImageName) RepoName() string {
	image := d.Image
	if d.Org == "library" {
		return image
	}
	return fmt.Sprintf("%s/%s", d.Org, image)
}
func (p dockerhubImage) Platform() string {
	if p.Variant != "" {
		return fmt.Sprintf("%s/%s/%s", p.Os, p.Architecture, p.Variant)
	}
	return fmt.Sprintf("%s/%s", p.Os, p.Architecture)
}

func parsePlatform(platform string) *dockerImagePlatform {
	if platform == "" {
		return nil
	}
	var p dockerImagePlatform
	split := strings.Split(platform, "/")
	if len(split) == 1 {
		p.Os = platform
		return &p
	}
	p.Os = split[0]
	if len(split) >= 1 {
		p.Architecture = split[1]
		if len(split) == 3 {
			p.Variant = split[2]
		}
	}
	return &p
}

func (i dockerhubImage) IsPlatform(p *dockerImagePlatform) bool {
	if i.Os != p.Os {
		return false
	}
	if p.Architecture != "" && i.Architecture != p.Architecture {
		return false
	}
	if p.Variant != "" && i.Variant != p.Variant {
		return false
	}
	return true
}

func parseDockerImage(imageName string) dockerImageName {
	var org, image, tag string

	if strings.Contains(imageName, ":") {
		split := strings.Split(imageName, ":")
		imageName = split[0]
		tag = split[1]
	}

	// default org is "library"
	if strings.Contains(imageName, "/") {
		split := strings.Split(imageName, "/")
		org = split[0]
		image = split[1]
	} else {
		org = "library"
		image = imageName
	}

	return dockerImageName{
		Org:   org,
		Image: image,
		Tag:   tag,
	}
}

func getDockerTagImagesDetails(image dockerImageName) ([]dockerhubImage, error) {
	var images []dockerhubImage
	res, err := http.Get(fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/%s", image, image.Tag))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var dResponse dockerhubTag
	err = json.NewDecoder(res.Body).Decode(&dResponse)
	if err != nil {
		return nil, err
	}

	for _, i := range dResponse.Images {
		i.FullName = image
		if i.Os != "unknown" {
			images = append(images, i)
		}
	}

	return images, nil
}

func getDockerTagsImages(image dockerImageName) ([]dockerhubImage, error) {
	var images []dockerhubImage
	res, err := http.Get(fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/?page_size=25&page=1&ordering=last_updated", image))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var dResponse dockerhubResponse
	err = json.NewDecoder(res.Body).Decode(&dResponse)
	if err != nil {
		return nil, err
	}

	if dResponse.Count == 0 || dResponse.Results == nil {
		return nil, fmt.Errorf("no tags found")
	}

	for _, t := range dResponse.Results {
		for _, i := range t.Images {
			i.FullName = image
			i.FullName.Tag = t.Name

			if i.Os != "unknown" {
				images = append(images, i)
			}
		}
	}

	return images, nil
}

func (DockerHub) Parse(options *parser.Options) (*feeds.Feed, error) {
	var feed feeds.Feed
	imageNameStr := options.Get("image").(string)
	if imageNameStr[0] == '/' {
		imageNameStr = imageNameStr[1:]
	}
	imageName := parseDockerImage(imageNameStr)
	var platform *dockerImagePlatform
	if options.Get("platform") == nil {
		platform = nil
	} else {
		platform = parsePlatform(options.Get("platform").(string))
	}

	var images []dockerhubImage
	var err error

	if imageName.Tag != "" {
		images, err = getDockerTagImagesDetails(imageName)
		if err != nil {
			return nil, fmt.Errorf("image not found")
		}
	} else {
		images, err = getDockerTagsImages(imageName)
		if err != nil {
			return nil, fmt.Errorf("tag not found")
		}
	}

	if platform != nil {
		var filteredImages []dockerhubImage
		for _, i := range images {
			if i.IsPlatform(platform) {
				filteredImages = append(filteredImages, i)
			}
		}
		images = filteredImages
	}

	var lastPushed time.Time

	for _, i := range images {

		var item feeds.Item

		if i.FullName.Org != "library" {
			item.Author = &feeds.Author{
				Name: i.FullName.Org,
			}
		}
		imagePushed, err := time.Parse("2006-01-02T15:04:05.999999Z", i.LastPushed)
		if err != nil {
			fmt.Println("Error while parsing date :", err)
			continue
		}
		item.Title = fmt.Sprintf("%s %s", i.FullName.Pretty(), i.Platform())
		item.Content = fmt.Sprintf("The %s image %s was pushed on %v", i.FullName.Pretty(), i.Platform(), i.LastPushed)
		item.Description = item.Content
		item.Link = &feeds.Link{Href: i.FullName.GetImageURL(i.Digest)}
		item.Id = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(i.FullName.Pretty(), i.FullName, i.Os, i.Architecture, i.LastPushed))))
		item.Updated = imagePushed
		item.Created = imagePushed

		if lastPushed.Before(imagePushed) {
			lastPushed = imagePushed
		}

		feed.Items = append(feed.Items, &item)
	}
	feed.Title = fmt.Sprintf("%s Images", imageName.Pretty())
	feed.Description = fmt.Sprintf("The latest %s images", imageName.Pretty())

	feed.Link = &feeds.Link{Href: imageName.GetUiURL()}
	return &feed, nil
}

type DockerHub struct{}

func DockerHubParser() parser.Parser {
	return DockerHub{}
}
