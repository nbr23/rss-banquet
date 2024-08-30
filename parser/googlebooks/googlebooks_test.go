package googlebooks

import (
	"fmt"
	"testing"
	"time"

	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

// Amélie Nothomb has been publishing yearly for 30 years. Don't break my tests!
func TestAmelieNothomb(t *testing.T) {
	options := Googlebooks{}.GetOptions()
	options.Set("author", "Amélie Nothomb")
	options.Set("language", "fr")
	options.Set("year-min", fmt.Sprintf("%d", time.Now().Year()-1))
	testsuite.TestParseSuccess(
		t,
		Googlebooks{},
		&options,
		2,
		`^.* - Amélie Nothomb$`,
	)
}
