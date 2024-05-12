package dockerhub

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/utils"
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
					Value: "nbr23/rss-banquet",
				},
			},
			Parser: DockerHub{},
		},
		1,
		`^nbr23/rss-banquet:[-\d\w]+ linux/[\d\w]+$`,
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
					Value: "nbr23/rss-banquet",
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
		`^nbr23/rss-banquet:[-\d\w]+ linux/arm64+$`,
	)
}

func TestParseDockerImage(t *testing.T) {

	testCases := []struct {
		name  string
		image dockerImageName
	}{
		{"alpine", dockerImageName{Org: "library", Image: "alpine", Tag: ""}},
		{"alpine:latest", dockerImageName{Org: "library", Image: "alpine", Tag: "latest"}},
		{"nbr23/rss-banquet", dockerImageName{Org: "nbr23", Image: "rss-banquet", Tag: ""}},
		{"nbr23/rss-banquet:latest", dockerImageName{Org: "nbr23", Image: "rss-banquet", Tag: "latest"}},
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
		{"nbr23/rss-banquet", "nbr23/rss-banquet"},
		{"nbr23/rss-banquet:latest", "nbr23/rss-banquet"},
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
		"nbr23/rss-banquet",
		"nbr23/rss-banquet:latest",
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
