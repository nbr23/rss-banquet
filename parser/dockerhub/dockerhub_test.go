package dockerhub

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

func TestDockerHubParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		DockerHub{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{
					Flag:  "image",
					Type:  "string",
					Value: "nbr23/youtube-dl-server",
				},
			},
			Parser: DockerHub{},
		},
		1,
		`^nbr23/youtube-dl-server:[-\d\w]+ linux/[\d\w]+$`,
	)
}

func TestDockerHubParsePlatform(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		DockerHub{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{
					Flag:  "image",
					Type:  "string",
					Value: "nbr23/youtube-dl-server",
				},
				&parser.Option{
					Flag:  "platform",
					Type:  "string",
					Value: "linux/arm64",
				},
			},
			Parser: DockerHub{},
		},
		1,
		`^nbr23/youtube-dl-server:[-\d\w]+ linux/arm64+$`,
	)
}

func TestParseDockerImage(t *testing.T) {

	testCases := []struct {
		name  string
		image dockerImageName
	}{
		{"alpine", dockerImageName{Org: "library", Image: "alpine", Tag: ""}},
		{"alpine:latest", dockerImageName{Org: "library", Image: "alpine", Tag: "latest"}},
		{"nbr23/youtube-dl-server", dockerImageName{Org: "nbr23", Image: "youtube-dl-server", Tag: ""}},
		{"nbr23/youtube-dl-server:latest", dockerImageName{Org: "nbr23", Image: "youtube-dl-server", Tag: "latest"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			image := parseDockerImage(tc.name)
			if image != tc.image {
				t.Errorf("got %q, wanted %q", image, tc.image)
			}
		})
	}
}

func TestDockerImageString(t *testing.T) {

	testCases := []struct {
		name     string
		fullName string
	}{
		{"alpine", "library/alpine"},
		{"alpine:latest", "library/alpine"},
		{"nbr23/youtube-dl-server", "nbr23/youtube-dl-server"},
		{"nbr23/youtube-dl-server:latest", "nbr23/youtube-dl-server"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			image := parseDockerImage(tc.name).String()
			if image != tc.fullName {
				t.Errorf("got %q, wanted %q", image, tc.fullName)
			}
		})
	}
}

func TestDockerImagePretty(t *testing.T) {

	testCases := []string{
		"alpine",
		"alpine:latest",
		"nbr23/youtube-dl-server",
		"nbr23/youtube-dl-server:latest",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			image := parseDockerImage(tc).Pretty()
			if image != tc {
				t.Errorf("got %q, wanted %q", image, tc)
			}
		})
	}
}
