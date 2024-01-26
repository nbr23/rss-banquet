package psupdates

import (
	"testing"

	testsuite "github.com/nbr23/atomic-banquet/utils"
)

func TestPS5UpdatesParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		PSUpdates{},
		map[string]interface{}{"hardware": "ps5", "local": "en-us"},
		1,
		`^PS5 Update: [^\s]+.*$`,
	)
}

func TestPS4UpdatesParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		PSUpdates{},
		map[string]interface{}{"hardware": "ps4", "local": "en-us"},
		1,
		`^PS4 Update: [^\s]+.*$`,
	)
}

func TestPSUpdatesBadOptions(t *testing.T) {
	testsuite.TestParseFailure(
		t,
		PSUpdates{},
		map[string]interface{}{"hardware": "ps1", "local": "en-us"},
	)
}
