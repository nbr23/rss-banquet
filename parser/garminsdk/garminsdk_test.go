package garminsdk

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

func TestGarminSDKParseFIT(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		GarminSDK{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{
					Flag:  "sdks",
					Type:  "stringSlice",
					Value: "fit",
				},
			},
			Parser: GarminSDK{},
		},
		1,
		`^\[\w+ \d+, \d+\] Garmin fit SDK Update: [^\s]+.*$`,
		``,
	)
}

func TestGarminSDKParseConnectIQ(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		GarminSDK{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{
					Flag:  "sdks",
					Type:  "stringSlice",
					Value: "connect-iq",
				},
			},
			Parser: GarminSDK{},
		},
		1,
		`^\[\w+ \d+, \d+\] Garmin connect-iq SDK Update: [^\s]+.*$`,
		``,
	)
}
