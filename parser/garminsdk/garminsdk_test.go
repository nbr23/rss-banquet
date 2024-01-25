package garminsdk

import (
	"regexp"
	"testing"
)

func TestGarminSDKParseFIT(t *testing.T) {
	psu := GarminSDK{}
	parsed, err := psu.Parse(map[string]interface{}{"sdks": []interface{}{"fit"}})
	if err != nil {
		t.Errorf("Unable to parse GarminSDK FIT: %s", err)
	}

	if parsed.Items == nil || len(parsed.Items) == 0 {
		t.Errorf("Unable to parse GarminSDK FIT: no items in feed")
	}

	if parsed.Items[0].Title == "" {
		t.Errorf("Unable to parse GarminSDK FIT: no title in feed item")
	}

	r := regexp.MustCompile(`^\[\w+ \d+, \d+\] Garmin fit SDK Update: [^\s]+.*$`)
	if !r.MatchString(parsed.Items[0].Title) {
		t.Errorf("Unable to parse GarminSDK FIT: title doesn't match expected format")
	}
}

func TestGarminSDKParseConnectIQ(t *testing.T) {
	psu := GarminSDK{}
	parsed, err := psu.Parse(map[string]interface{}{"sdks": []interface{}{"connect-iq"}})
	if err != nil {
		t.Errorf("Unable to parse GarminSDK ConnectIQ: %s", err)
	}

	if parsed.Items == nil || len(parsed.Items) == 0 {
		t.Errorf("Unable to parse GarminSDK ConnectIQ: no items in feed")
	}

	if parsed.Items[0].Title == "" {
		t.Errorf("Unable to parse GarminSDK ConnectIQ: no title in feed item")
	}

	r := regexp.MustCompile(`^\[\w+ \d+, \d+\] Garmin connect-iq SDK Update: [^\s]+.*$`)
	if !r.MatchString(parsed.Items[0].Title) {
		t.Errorf("Unable to parse GarminSDK ConnectIQ: title doesn't match expected format, got '%s'", parsed.Items[0].Title)
	}
}
