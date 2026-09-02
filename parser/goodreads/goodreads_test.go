package goodreads

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nbr23/rss-banquet/config"
	"github.com/nbr23/rss-banquet/parser"
	"github.com/nbr23/rss-banquet/testsuite"
	"github.com/stretchr/testify/assert"
)

// Amélie Nothomb has been publishing yearly for 30 years. Don't break my tests!
func TestAmelieNothomb(t *testing.T) {
	options := GoodReads{}.GetOptions()
	options.Set("authorId", "40416.Am_lie_Nothomb")
	options.Set("language", "fr")
	options.Set("year-min", fmt.Sprintf("%d", time.Now().Year()-1))
	options.Set("bookFormats", "paperback,hardcover,kindle")
	testsuite.TestParseSuccess(
		t,
		GoodReads{},
		&options,
		1,
		`^.* - Amélie Nothomb.*$`,
		`^Books by Amélie Nothomb - French$`,
	)
}

func TestHttpGetRetryBotChallenge(t *testing.T) {
	config.InitConfig()
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("x-amzn-waf-action", "challenge")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	resp, err := httpGetRetry(ts.URL)

	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, parser.ErrUpstreamBlocked))
	assert.Equal(t, 1, calls)
}

func TestHttpGetRetryTransientFailure(t *testing.T) {
	config.InitConfig()
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	resp, err := httpGetRetry(ts.URL)

	assert.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "ok", string(body))
	assert.Equal(t, 2, calls)
}

func TestGetBooksListChallengedBookPage(t *testing.T) {
	config.InitConfig()
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			fmt.Fprintf(w, `<html><body><h1>Books by Test Author</h1>
				<div itemtype="http://schema.org/Book">
					<a itemprop="url" href="%s/book/show/1-test">Test Book</a>
					<span>published 2025</span>
				</div></body></html>`, ts.URL)
			return
		}
		w.Header().Set("x-amzn-waf-action", "challenge")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	books, _, err := getBooksList(ts.URL, "French", 2025, []string{"paperback"})

	assert.Empty(t, books)
	assert.True(t, errors.Is(err, parser.ErrUpstreamBlocked))
}

func TestHttpGetRetryNotFound(t *testing.T) {
	config.InitConfig()
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	resp, err := httpGetRetry(ts.URL)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, parser.ErrUpstreamBlocked))
	assert.Equal(t, 1, calls)
}
