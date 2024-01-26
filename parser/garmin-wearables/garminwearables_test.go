package garminwearables

import (
	"testing"

	testsuite "github.com/nbr23/atomic-banquet/utils"
)

func TestGarminWearablesParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		GarminWearables{},
		map[string]interface{}{},
		1,
		`^\[[\w]+\] Garmin Wearable Update$`,
	)
}
