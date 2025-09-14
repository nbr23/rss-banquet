package googlebooks

import (
	"testing"

	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

// Amélie Nothomb has been publishing yearly for 30 years. Don't break my tests!
func TestAmelieNothomb(t *testing.T) {
	options := Googlebooks{}.GetOptions()
	options.Set("author", "Amélie Nothomb")
	options.Set("language", "fr")
	testsuite.TestParseSuccess(
		t,
		Googlebooks{},
		&options,
		1,
		`^.* - Amélie Nothomb.*$`,
		``,
	)
}
