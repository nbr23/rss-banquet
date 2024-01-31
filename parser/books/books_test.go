package books

import (
	"testing"

	testsuite "github.com/nbr23/atomic-banquet/utils"
)

// Amélie Nothomb has been publishing yearly for 30 years. Don't break my tests!
func TestAmelieNothomb(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		Books{},
		map[string]interface{}{"author": "Amélie Nothomb", "language": "fr"},
		1,
		`^.* - Amélie Nothomb$`,
	)
}
