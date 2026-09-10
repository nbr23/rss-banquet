package goodreads

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nbr23/rss-banquet/config"
	"github.com/nbr23/rss-banquet/parser"
	"github.com/nbr23/rss-banquet/testsuite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type goodreadsTestTransport func(*http.Request) (*http.Response, error)

func (transport goodreadsTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport(req)
}

func mockGoodreadsPages(t *testing.T, pages map[string]string) {
	t.Helper()
	config.InitConfig()
	previousTransport := http.DefaultTransport
	bookDetailsCacheMu.Lock()
	previousCache := bookDetailsCache
	bookDetailsCache = make(map[string]cachedBook)
	bookDetailsCacheMu.Unlock()
	t.Cleanup(func() {
		http.DefaultTransport = previousTransport
		bookDetailsCacheMu.Lock()
		bookDetailsCache = previousCache
		bookDetailsCacheMu.Unlock()
	})
	http.DefaultTransport = goodreadsTestTransport(func(req *http.Request) (*http.Response, error) {
		body, ok := pages[req.URL.Path]
		status := http.StatusOK
		if !ok || req.URL.Host != "www.goodreads.com" {
			t.Errorf("unexpected request: %s", req.URL)
			status = http.StatusNotFound
		}
		return &http.Response{
			StatusCode: status,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
}

func goodreadsListRow(bookPath, workPath string) string {
	return fmt.Sprintf(`<div itemtype="http://schema.org/Book">
		<a itemprop="url" href="%s">Test Book</a>
		<a href="%s">Editions</a>
		</div>`, bookPath, workPath)
}

func goodreadsEdition(bookPath, date string) string {
	return fmt.Sprintf(`<div class="editionData">
		<a class="bookTitle" href="%s">Test Book</a>
		<div class="dataRow">Paperback, Published %s</div>
		<div class="dataRow"><div class="dataTitle">Edition language:</div>
		<div class="dataValue">English</div></div>
		</div>`, bookPath, date)
}

func goodreadsBookPage(date string, prerelease bool) string {
	page := fmt.Sprintf(`<h1 data-testid="bookTitle">Test Book</h1>
		<div class="BookPageMetadataSection__contributor">Test Author</div>
		<p data-testid="pagesFormat">Paperback</p>
		<p data-testid="publicationInfo">Published %s</p>
		<script type="application/ld+json">{"inLanguage":"English"}</script>`, date)
	if prerelease {
		page += fmt.Sprintf(`<div class="PreReleaseDetails">Expected publication %s</div>`, date)
	}
	return page
}

func TestStableWorkIdentityAcrossPolls(t *testing.T) {
	for _, option := range []string{"authorId", "seriesId"} {
		t.Run(option, func(t *testing.T) {
			pages := make(map[string]string)
			mockGoodreadsPages(t, pages)
			options := GoodReads{}.GetOptions()
			options.Set(option, "1-test")
			options.Set("year-min", "2000")
			listPath := "/author/list/1-test"
			if option == "seriesId" {
				listPath = "/series/1-test"
			}
			futureYear := time.Now().Year() + 2
			polls := []struct {
				bookID     string
				date       string
				prerelease bool
			}{
				{"245835653", fmt.Sprintf("January 2 %d", futureYear), true},
				{"252160570", fmt.Sprintf("February 3 %d", futureYear), true},
				{"252160570", "March 4 2001", false},
				{"245645195", "March 4 2001", false},
			}
			for _, poll := range polls {
				bookPath := "/book/show/" + poll.bookID + "-test"
				workPath := "/work/editions/123-test?sort=default"
				pages[listPath] = "<h1>Test Books</h1>" + strings.Repeat(goodreadsListRow(bookPath, workPath), 2)
				pages["/work/editions/123-test"] = goodreadsEdition(bookPath, poll.date)
				pages[bookPath] = goodreadsBookPage(poll.date, poll.prerelease)

				feed, err := (GoodReads{}).Parse(&options)

				require.NoError(t, err)
				require.Len(t, feed.Items, 1)
				assert.Equal(t, "goodreads:work:123", feed.Items[0].Id)
				assert.Equal(t, "https://www.goodreads.com"+bookPath, feed.Items[0].Link.Href)
				assert.Contains(t, feed.Items[0].Content, poll.date)
				for _, serialize := range []func() (string, error){feed.ToRss, feed.ToAtom, feed.ToJSON} {
					output, err := serialize()
					require.NoError(t, err)
					assert.Equal(t, 1, strings.Count(output, "goodreads:work:123"), output)
				}
			}
		})
	}
}

func TestEditionSelectionIgnoresPageOrder(t *testing.T) {
	date := fmt.Sprintf("January 2 %d", time.Now().Year()+2)
	laterDate := fmt.Sprintf("January 3 %d", time.Now().Year()+2)
	pages := map[string]string{
		"/author/list/1-test": "<h1>Test Books</h1>" + goodreadsListRow("/book/show/10-test", "/work/editions/123-test"),
		"/book/show/8-test":   goodreadsBookPage(laterDate, true),
		"/book/show/9-test":   goodreadsBookPage(date, true),
		"/book/show/9-z":      goodreadsBookPage(date, true),
		"/book/show/10-test":  goodreadsBookPage(date, true),
	}
	mockGoodreadsPages(t, pages)
	for _, editions := range [][]string{
		{"9-test", "10-test", "8-test"},
		{"8-test", "10-test", "9-test"},
		{"9-z", "9-test"},
		{"9-test", "9-z"},
	} {
		pages["/work/editions/123-test"] = ""
		for _, edition := range editions {
			publicationDate := date
			if edition == "8-test" {
				publicationDate = laterDate
			}
			pages["/work/editions/123-test"] += goodreadsEdition("/book/show/"+edition, publicationDate)
		}

		books, _, err := getBooksList("https://www.goodreads.com/author/list/1-test", "English", 2000, []string{"paperback"})

		require.NoError(t, err)
		require.Len(t, books, 1)
		assert.Equal(t, "https://www.goodreads.com/book/show/9-test", books[0].Link)
	}
}

func TestBookDeduplicationAfterFiltering(t *testing.T) {
	pages := map[string]string{
		"/author/list/1-test": "<h1>Test Books</h1>" +
			goodreadsListRow("/book/show/1-test", "/work/editions/123-test") +
			goodreadsListRow("/book/show/2-test", "/work/editions/123-test") +
			goodreadsListRow("/book/show/3-test", "/work/editions/123-test") +
			goodreadsListRow("/book/show/4-test", "/work/editions/456-test") +
			goodreadsListRow("/book/show/5-test?from=list", "") +
			goodreadsListRow("/book/show/5.changed-slug", ""),
		"/work/editions/123-test":   "",
		"/work/editions/456-test":   "",
		"/book/show/1-test":         strings.ReplaceAll(goodreadsBookPage("January 2 2001", false), "English", "French"),
		"/book/show/2-test":         goodreadsBookPage("January 2 2001", false),
		"/book/show/3-test":         goodreadsBookPage("January 2 2001", false),
		"/book/show/4-test":         goodreadsBookPage("January 2 2001", false),
		"/book/show/5-test":         goodreadsBookPage("January 2 2001", false),
		"/book/show/5.changed-slug": goodreadsBookPage("January 2 2001", false),
	}
	mockGoodreadsPages(t, pages)

	books, _, err := getBooksList("https://www.goodreads.com/author/list/1-test", "English", 2000, []string{"paperback"})

	require.NoError(t, err)
	require.Len(t, books, 3)
	assert.Equal(t, "https://www.goodreads.com/book/show/2-test", books[0].Link)
	assert.Equal(t, "goodreads:work:123", books[0].feedID())
	assert.Equal(t, "goodreads:work:456", books[1].feedID())
	assert.Equal(t, "goodreads:book:5", books[2].feedID())
}

func TestCachedBookWorkIdentityIsRequestLocal(t *testing.T) {
	pages := map[string]string{"/book/show/1-test": goodreadsBookPage("January 2 2001", false)}
	mockGoodreadsPages(t, pages)
	link := "https://www.goodreads.com/book/show/1-test"
	for _, workID := range []string{"", "123", "456", ""} {
		book, err := getBookDetailsCached(&GRBook{Link: link, WorkID: workID})

		require.NoError(t, err)
		assert.Equal(t, workID, book.WorkID)
		assert.Equal(t, "Test Book", book.Title)
		book.Title = "Changed by caller"
		book.WorkID = "789"
		delete(pages, "/book/show/1-test")
	}
}

func TestBookIdentityFallback(t *testing.T) {
	for _, test := range []struct {
		link string
		want string
	}{
		{"https://www.goodreads.com/book/show/123-title?from=list#details", "goodreads:book:123"},
		{"/book/show/123.title", "goodreads:book:123"},
		{"/book/show/123", "goodreads:book:123"},
		{"/book/show/000123-title", "goodreads:book:123"},
		{"/book/show/123invalid", "/book/show/123invalid"},
		{"/book/show/title", "/book/show/title"},
		{"/other?next=/book/show/123", "/other?next=/book/show/123"},
	} {
		t.Run(test.link, func(t *testing.T) {
			book := GRBook{Link: test.link, PublicationDate: "January 2 2001"}
			assert.Equal(t, test.want, book.feedID())
			book.PublicationDate = "February 3 2002"
			assert.Equal(t, test.want, book.feedID())
		})
	}
}
