package garminsdk

import (
	"testing"

	testsuite "github.com/nbr23/atomic-banquet/utils"
)

func TestGarminSDKParseFIT(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		GarminSDK{},
		map[string]interface{}{"sdks": []string{"fit"}},
		1,
		`^\[\w+ \d+, \d+\] Garmin fit SDK Update: [^\s]+.*$`,
	)
}

func TestGarminSDKParseConnectIQ(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		GarminSDK{},
		map[string]interface{}{"sdks": []string{"connect-iq"}},
		1,
		`^\[\w+ \d+, \d+\] Garmin connect-iq SDK Update: [^\s]+.*$`,
	)
}
