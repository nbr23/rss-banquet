package imdb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetArtistWorks(t *testing.T) {
	response := `{
		"data": {
			"name": {
				"nameText": {"text": "Test Artist"},
				"credits": {
					"edges": [
						{
							"node": {
								"title": {
									"id": "tt1234567",
									"titleText": {"text": "Test Movie"},
									"releaseYear": {"year": 2023},
									"releaseDate": {"day": 15, "month": 6, "year": 2023},
									"titleType": {"text": "Movie"},
									"primaryImage": {"url": "https://m.media-amazon.com/images/M/poster.jpg"}
								},
								"category": {"text": "Actor"},
								"characters": [{"name": "John Doe"}]
							}
						},
						{
							"node": {
								"title": {
									"id": "tt7654321",
									"titleText": {"text": "Test Show"},
									"releaseYear": {"year": 2022},
									"releaseDate": null,
									"titleType": {"text": "TV Series"}
								},
								"category": {"text": "Producer"}
							}
						}
					]
				}
			}
		}
	}`

	var capturedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	defer ts.Close()

	post := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return http.Post(ts.URL, contentType, body)
	}

	works, artistName, err := getArtistWorks("nm0000000", 25, post)
	assert.NoError(t, err)
	assert.Equal(t, "Test Artist", artistName)
	assert.Len(t, works, 2)

	assert.Equal(t, "Test Movie", works[0].Title)
	assert.Equal(t, 2023, works[0].Year)
	assert.Equal(t, "John Doe", works[0].Role)
	assert.Equal(t, "Actor", works[0].Category)
	assert.Equal(t, "Movie", works[0].TitleType)
	assert.True(t, works[0].HasFullDate)
	assert.Equal(t, "2023-06-15", works[0].ReleaseDate.Format("2006-01-02"))
	assert.Contains(t, works[0].Link, "/title/tt1234567/")
	assert.Equal(t, "https://m.media-amazon.com/images/M/poster.jpg", works[0].ImageURL)
	assert.Equal(t, "", works[1].ImageURL)

	assert.Equal(t, "Test Show", works[1].Title)
	assert.Equal(t, "Producer", works[1].Category)
	assert.False(t, works[1].HasFullDate)
	assert.Equal(t, 2022, works[1].ReleaseDate.Year())

	vars, _ := capturedBody["variables"].(map[string]any)
	assert.Equal(t, "nm0000000", vars["id"])
	assert.EqualValues(t, 25, vars["first"])
	assert.Contains(t, capturedBody["query"], "NameCredits")
}

func TestGetArtistWorksGraphqlError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"errors":[{"message":"PersistedQueryNotFound"}]}`))
	}))
	defer ts.Close()

	post := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return http.Post(ts.URL, contentType, body)
	}

	_, _, err := getArtistWorks("nm0000000", 25, post)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "PersistedQueryNotFound"))
}

func TestGetArtistNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"name":{"nameText":{"text":""},"credits":{"edges":[]}}}}`))
	}))
	defer ts.Close()

	post := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return http.Post(ts.URL, contentType, body)
	}

	_, _, err := getArtistWorks("nm9999999", 25, post)
	assert.Error(t, err)
}
