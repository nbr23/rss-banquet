package lego

import (
	"testing"

	testsuite "github.com/nbr23/atomic-banquet/utils"
)

func TestLegoParseNew(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		Lego{},
		map[string]interface{}{"category": "new"},
		1,
		`^\[\w+\] [0-9]+ - .* - (Available now|Pre-order this item today,) .*$`,
	)
}

func TestLegoParseComingSoon(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		Lego{},
		map[string]interface{}{"category": "coming-soon"},
		1,
		`^\[\w+\] [0-9]+ - .* - Coming Soon .*$`,
	)
}
