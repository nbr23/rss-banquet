package psupdates

import (
	"regexp"
	"testing"
)

func TestPS5UpdatesParse(t *testing.T) {
	psu := PSUpdates{}
	parsed, err := psu.Parse(map[string]interface{}{"hardware": "ps5", "local": "en-us"})
	if err != nil {
		t.Errorf("Unable to parse PSUpdates: %s", err)
	}

	if parsed.Items == nil || len(parsed.Items) == 0 {
		t.Errorf("Unable to parse PSUpdates: no items in feed")
	}

	if parsed.Items[0].Title == "" {
		t.Errorf("Unable to parse PSUpdates: no title in feed item")
	}

	r := regexp.MustCompile(`^PS5 Update: [^\s]+.*$`)
	if !r.MatchString(parsed.Items[0].Title) {
		t.Errorf("Unable to parse PSUpdates, title doesn't match expected format, got '%s'", parsed.Items[0].Title)
	}
}

func TestPS4UpdatesParse(t *testing.T) {
	psu := PSUpdates{}
	parsed, err := psu.Parse(map[string]interface{}{"hardware": "ps4", "local": "en-us"})
	if err != nil {
		t.Errorf("Unable to parse PSUpdates: %s", err)
	}

	if parsed.Items == nil || len(parsed.Items) == 0 {
		t.Errorf("Unable to parse PSUpdates: no items in feed")
	}

	if parsed.Items[0].Title == "" {
		t.Errorf("Unable to parse PSUpdates: no title in feed item")
	}

	r := regexp.MustCompile(`^PS4 Update: [^\s]+.*$`)
	if !r.MatchString(parsed.Items[0].Title) {
		t.Errorf("Unable to parse PSUpdates, title doesn't match expected format, got '%s'", parsed.Items[0].Title)
	}
}

func TestPSUpdatesBadOptions(t *testing.T) {
	psu := PSUpdates{}
	_, err := psu.Parse(map[string]interface{}{"hardware": "ps1", "local": "en-us"})
	if err == nil || err.Error() != "unable to fetch the update page, status code: 404" {
		t.Errorf("Failed to fail on bad options: %s", err)
		return
	}
}
