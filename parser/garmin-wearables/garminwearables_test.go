package garminwearables

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/utils"
)

func TestGarminWearablesParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		GarminWearables{},
		&parser.Options{},
		1,
		`^\[[\w]+\] Garmin Wearable Update$`,
	)
}
