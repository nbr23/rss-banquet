package dockerhub

import (
	"testing"

	testsuite "github.com/nbr23/atomic-banquet/utils"
)

func TestDockerHubParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		DockerHub{},
		map[string]interface{}{"image": "nbr23/atomic-banquet"},
		1,
		`^nbr23/atomic-banquet:[-\d\w]+ linux/[\d\w]+$`,
	)
}

func TestDockerHubParseArch(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		DockerHub{},
		map[string]interface{}{"image": "nbr23/atomic-banquet", "arch": "linux/arm64"},
		1,
		`^nbr23/atomic-banquet:[-\d\w]+ linux/arm64+$`,
	)
}

func TestParseDockerImage(t *testing.T) {

	testCases := []struct {
		name  string
		image dockerImageName
	}{
		{"alpine", dockerImageName{Org: "library", Image: "alpine", Tag: ""}},
		{"alpine:latest", dockerImageName{Org: "library", Image: "alpine", Tag: "latest"}},
		{"nbr23/atomic-banquet", dockerImageName{Org: "nbr23", Image: "atomic-banquet", Tag: ""}},
		{"nbr23/atomic-banquet:latest", dockerImageName{Org: "nbr23", Image: "atomic-banquet", Tag: "latest"}},
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
		{"nbr23/atomic-banquet", "nbr23/atomic-banquet"},
		{"nbr23/atomic-banquet:latest", "nbr23/atomic-banquet"},
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
		"nbr23/atomic-banquet",
		"nbr23/atomic-banquet:latest",
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
