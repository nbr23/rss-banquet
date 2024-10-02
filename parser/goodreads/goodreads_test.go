package goodreads

import (
	"fmt"
	"testing"
	"time"

	"github.com/nbr23/rss-banquet/testsuite"
)

// Amélie Nothomb has been publishing yearly for 30 years. Don't break my tests!
func TestAmelieNothomb(t *testing.T) {
	options := GoodReads{}.GetOptions()
	options.Set("authorId", "40416.Am_lie_Nothomb")
	options.Set("language", "fr")
	options.Set("year-min", fmt.Sprintf("%d", time.Now().Year()-1))
	testsuite.TestParseSuccess(
		t,
		GoodReads{},
		&options,
		1,
		`^.* - Amélie Nothomb.*$`,
		`^Books by Amélie Nothomb - French$`,
	)
}
