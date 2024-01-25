package psupdates

import (
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
}

func TestPSUpdatesBadOptions(t *testing.T) {
	psu := PSUpdates{}
	_, err := psu.Parse(map[string]interface{}{"hardware": "ps1", "local": "en-us"})
	if err == nil || err.Error() != "unable to fetch the update page, status code: 404" {
		t.Errorf("Failed to fail on bad options: %s", err)
		return
	}
}
