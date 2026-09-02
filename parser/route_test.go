package parser

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	utls "github.com/refraction-networking/utls"
	"github.com/stretchr/testify/assert"

	"github.com/nbr23/rss-banquet/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testFeed() *feeds.Feed {
	return &feeds.Feed{
		Title:   "Feed Title",
		Link:    &feeds.Link{Href: "https://example.com"},
		Created: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Items: []*feeds.Item{
			{
				Title:       "Item",
				Link:        &feeds.Link{Href: "https://example.com/1"},
				Description: "desc",
				Created:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

func serveFeedFormat(format string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/feed/stub?feedFormat=%s", format), nil)
	ServeFeed(c, testFeed())
	return w
}

func TestServeFeedRss(t *testing.T) {
	for _, format := range []string{"rss", "", "unknown"} {
		w := serveFeedFormat(format)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), `<?xml-stylesheet type="text/xsl" href="/rss-style.xsl"?>`)
		assert.Contains(t, w.Body.String(), "<title>Feed Title</title>")
	}
}

func TestServeFeedAtom(t *testing.T) {
	w := serveFeedFormat("atom")
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `<?xml-stylesheet type="text/xsl" href="/atom-style.xsl"?>`)
}

func TestServeFeedJson(t *testing.T) {
	w := serveFeedFormat("json")
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `"title": "Feed Title"`)
}

func TestServeFeedText(t *testing.T) {
	w := serveFeedFormat("text")
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, FeedToText(testFeed()), w.Body.String())
}

func TestRouteBuildsPath(t *testing.T) {
	tests := []struct {
		name   string
		parser *stubParser
		want   string
	}{
		{
			name:   "No required options",
			parser: &stubParser{name: "plain"},
			want:   "/feed/plain",
		},
		{
			name: "Required option becomes a path param",
			parser: &stubParser{
				name: "required",
				options: OptionsList{
					{Flag: "org", Type: "string", Required: true},
					{Flag: "image", Type: "string", Required: true},
					{Flag: "optional", Type: "string"},
				},
			},
			want: "/feed/required/:org/:image",
		},
		{
			name: "Required path option becomes a wildcard",
			parser: &stubParser{
				name: "wildcard",
				options: OptionsList{
					{Flag: "url", Type: "string", Required: true, IsPath: true},
				},
			},
			want: "/feed/wildcard/*url",
		},
		{
			name: "Static options are skipped",
			parser: &stubParser{
				name: "static",
				options: OptionsList{
					{Flag: "hidden", Type: "string", Required: true, IsStatic: true},
				},
			},
			want: "/feed/static",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gin.New()
			Route(g, tt.parser, GetFullOptions(tt.parser))

			var paths []string
			for _, r := range g.Routes() {
				paths = append(paths, r.Path)
			}
			assert.Contains(t, paths, tt.want)
		})
	}
}

func TestRoutePassesParameters(t *testing.T) {
	p := &stubParser{
		name: "stub",
		options: OptionsList{
			{Flag: "org", Type: "string", Required: true},
			{Flag: "limit", Type: "int", Default: "10"},
		},
	}
	baseOptions := GetFullOptions(p)

	g := gin.New()
	Route(g, p, baseOptions)

	w := httptest.NewRecorder()
	g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/feed/stub/myorg?limit=5", nil))

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "myorg", p.got.Get("org"))
	assert.Equal(t, 5, p.got.Get("limit"))

	w = httptest.NewRecorder()
	g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/feed/stub/otherorg", nil))

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "otherorg", p.got.Get("org"))
	assert.Equal(t, 10, p.got.Get("limit"))

	for _, option := range baseOptions.OptionsList {
		assert.Nil(t, option.Value, option.Flag)
	}
}

func TestRouteErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantBody string
	}{
		{
			name:     "Not found error",
			err:      NewNotFoundError("no such thing"),
			wantCode: 404,
			wantBody: "NotFoundError: no such thing",
		},
		{
			name:     "Internal error",
			err:      NewInternalError("upstream broke"),
			wantCode: 500,
			wantBody: "InternalError: upstream broke",
		},
		{
			name:     "Generic error",
			err:      fmt.Errorf("some parsing issue"),
			wantCode: 500,
			wantBody: "error parsing feed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &stubParser{
				name: "stub",
				parse: func(*Options) (*feeds.Feed, error) {
					return nil, tt.err
				},
			}

			g := gin.New()
			Route(g, p, GetFullOptions(p))

			w := httptest.NewRecorder()
			g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/feed/stub", nil))

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestRouteSortsFeed(t *testing.T) {
	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	p := &stubParser{
		name: "stub",
		parse: func(*Options) (*feeds.Feed, error) {
			return &feeds.Feed{
				Title: "Feed Title",
				Link:  &feeds.Link{Href: "https://example.com"},
				Items: []*feeds.Item{
					{Title: "older", Link: &feeds.Link{Href: "https://example.com/1"}, Description: "old desc", Created: older},
					{Title: "newer", Link: &feeds.Link{Href: "https://example.com/2"}, Description: "new desc", Created: newer},
				},
			}, nil
		},
	}

	g := gin.New()
	Route(g, p, GetFullOptions(p))

	w := httptest.NewRecorder()
	g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/feed/stub?feedFormat=text", nil))

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "# Feed Title | https://example.com\n- newer\n\thttps://example.com/2\n\tnew desc\n- older\n\thttps://example.com/1\n\told desc", w.Body.String())
}

func TestHttpGetSendsHeaders(t *testing.T) {
	var got http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	resp, err := HttpGet(ts.URL, map[string]any{
		"headers": map[string]string{
			"Referer":         "https://example.com",
			"Accept-Language": "en-us",
		},
	})
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, config.GetConfigOption("USER_AGENT"), got.Get("User-Agent"))
	assert.Equal(t, "https://example.com", got.Get("Referer"))
	assert.Equal(t, "en-us", got.Get("Accept-Language"))
}

func TestHttpGetWithoutOptions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(418)
	}))
	defer ts.Close()

	resp, err := HttpGet(ts.URL, nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 418, resp.StatusCode)
}

func TestHttpGetInvalidUrl(t *testing.T) {
	resp, err := HttpGet("://not a url", nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestHttpGetUtls(t *testing.T) {
	var got http.Header
	var proto string
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		proto = r.Proto
		w.Write([]byte("ok"))
	}))
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	pool := x509.NewCertPool()
	pool.AddCert(ts.Certificate())

	original := utlsConfig
	utlsConfig = func(host string) *utls.Config {
		return &utls.Config{ServerName: host, RootCAs: pool}
	}
	t.Cleanup(func() { utlsConfig = original })

	resp, err := HttpGetUtls(ts.URL, map[string]any{
		"headers": map[string]string{"Referer": "https://example.com"},
	})
	assert.NoError(t, err)
	if !assert.NotNil(t, resp) {
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "HTTP/2.0", proto)
	assert.Equal(t, config.GetConfigOption("USER_AGENT"), got.Get("User-Agent"))
	assert.Equal(t, "https://example.com", got.Get("Referer"))
}

func TestHttpGetUtlsInvalidUrl(t *testing.T) {
	resp, err := HttpGetUtls("://not a url", nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetRemoteFileLastModified(t *testing.T) {
	lastModified := "Mon, 15 Jan 2024 10:30:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", lastModified)
	}))
	defer ts.Close()

	got, err := GetRemoteFileLastModified(ts.URL)
	assert.NoError(t, err)
	assert.True(t, got.Equal(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)), got)
}

func TestGetRemoteFileLastModifiedBadStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	_, err := GetRemoteFileLastModified(ts.URL)
	assert.EqualError(t, err, "unable to fetch the update file, status code: 404")
}

func TestGetRemoteFileLastModifiedMissingHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	_, err := GetRemoteFileLastModified(ts.URL)
	assert.Error(t, err)
}

func TestGetRemoteFileLastModifiedUnreachable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := ts.URL
	ts.Close()

	_, err := GetRemoteFileLastModified(url)
	assert.Error(t, err)
}
