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
									"titleType": {"id": "movie", "text": "Movie"},
									"primaryImage": {"url": "https://m.media-amazon.com/images/M/poster.jpg"}
								},
								"category": {"id": "actor", "text": "Actor"},
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
									"titleType": {"id": "tvSeries", "text": "TV Series"}
								},
								"category": {"id": "producer", "text": "Producer"}
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

	works, artistName, err := getArtistWorks("nm0000000", post)
	assert.NoError(t, err)
	assert.Equal(t, "Test Artist", artistName)
	assert.Len(t, works, 2)

	assert.Equal(t, "Test Movie", works[0].Title)
	assert.Equal(t, 2023, works[0].Year)
	assert.Len(t, works[0].Credits, 1)
	assert.Equal(t, "John Doe", works[0].Credits[0].Role)
	assert.Equal(t, "Actor", works[0].Credits[0].Category)
	assert.Equal(t, "Movie", works[0].TitleType)
	assert.True(t, works[0].HasFullDate)
	assert.Equal(t, "2023-06-15", works[0].ReleaseDate.Format("2006-01-02"))
	assert.Contains(t, works[0].Link, "/title/tt1234567/")
	assert.Equal(t, "https://m.media-amazon.com/images/M/poster.jpg", works[0].ImageURL)
	assert.Equal(t, "", works[1].ImageURL)
	assert.Equal(t, "movie", works[0].TitleTypeID)
	assert.Equal(t, "actor", works[0].Credits[0].CategoryID)
	assert.Equal(t, "tvSeries", works[1].TitleTypeID)
	assert.Equal(t, "producer", works[1].Credits[0].CategoryID)

	assert.Equal(t, "Test Show", works[1].Title)
	assert.Equal(t, "Producer", works[1].Credits[0].Category)
	assert.False(t, works[1].HasFullDate)
	assert.Equal(t, 2022, works[1].ReleaseDate.Year())

	vars, _ := capturedBody["variables"].(map[string]any)
	assert.Equal(t, "nm0000000", vars["id"])
	assert.EqualValues(t, creditsPageSize, vars["first"])
	assert.Contains(t, capturedBody["query"], "NameCredits")
}

func TestFilterWorks(t *testing.T) {
	works := []IMDBWork{
		{Title: "A", TitleTypeID: "movie", Credits: []IMDBCredit{{CategoryID: "actor"}}},
		{Title: "B", TitleTypeID: "tvSeries", Credits: []IMDBCredit{{CategoryID: "producer"}}},
		{Title: "C", TitleTypeID: "movie", Credits: []IMDBCredit{{CategoryID: "director"}}},
	}

	all := filterWorks(works, []string{""}, []string{""})
	assert.Len(t, all, 3)

	byType := filterWorks(works, []string{"movie"}, nil)
	assert.Len(t, byType, 2)
	assert.Equal(t, "A", byType[0].Title)
	assert.Equal(t, "C", byType[1].Title)

	caseInsensitive := filterWorks(works, []string{"TVSERIES"}, nil)
	assert.Len(t, caseInsensitive, 1)
	assert.Equal(t, "B", caseInsensitive[0].Title)

	byCategory := filterWorks(works, nil, []string{"actor", "director"})
	assert.Len(t, byCategory, 2)

	both := filterWorks(works, []string{"movie"}, []string{"director"})
	assert.Len(t, both, 1)
	assert.Equal(t, "C", both[0].Title)
}

func TestGetArtistWorksDedupesCredits(t *testing.T) {
	response := `{
		"data": {
			"name": {
				"nameText": {"text": "David Cronenberg"},
				"credits": {
					"edges": [
						{
							"node": {
								"title": {
									"id": "tt0099763",
									"titleText": {"text": "Naked Lunch"},
									"releaseYear": {"year": 1991},
									"titleType": {"id": "movie", "text": "Movie"}
								},
								"category": {"id": "director", "text": "Director"}
							}
						},
						{
							"node": {
								"title": {
									"id": "tt0099763",
									"titleText": {"text": "Naked Lunch"},
									"releaseYear": {"year": 1991},
									"titleType": {"id": "movie", "text": "Movie"}
								},
								"category": {"id": "writer", "text": "Writer"}
							}
						}
					]
				}
			}
		}
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
	}))
	defer ts.Close()

	post := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return http.Post(ts.URL, contentType, body)
	}

	works, _, err := getArtistWorks("nm0000343", post)
	assert.NoError(t, err)
	assert.Len(t, works, 1)
	assert.Len(t, works[0].Credits, 2)
	assert.Equal(t, "Director", works[0].Credits[0].Category)
	assert.Equal(t, "Writer", works[0].Credits[1].Category)

	byCategory := filterWorks(works, nil, []string{"writer"})
	assert.Len(t, byCategory, 1)
}

func TestGetArtistWorksPaginates(t *testing.T) {
	page1 := `{
		"data": {
			"name": {
				"nameText": {"text": "Test Artist"},
				"credits": {
					"pageInfo": {"hasNextPage": true, "endCursor": "CURSOR1"},
					"edges": [
						{"node": {"title": {"id": "tt0000001", "titleText": {"text": "First"}, "releaseYear": {"year": 2020}}, "category": {"id": "actor", "text": "Actor"}}}
					]
				}
			}
		}
	}`
	page2 := `{
		"data": {
			"name": {
				"nameText": {"text": "Test Artist"},
				"credits": {
					"pageInfo": {"hasNextPage": false, "endCursor": ""},
					"edges": [
						{"node": {"title": {"id": "tt0000002", "titleText": {"text": "Second"}, "releaseYear": {"year": 2021}}, "category": {"id": "director", "text": "Director"}}}
					]
				}
			}
		}
	}`

	var cursors []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		vars, _ := parsed["variables"].(map[string]any)
		after, _ := vars["after"].(string)
		cursors = append(cursors, after)
		if after == "CURSOR1" {
			w.Write([]byte(page2))
		} else {
			w.Write([]byte(page1))
		}
	}))
	defer ts.Close()

	post := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return http.Post(ts.URL, contentType, body)
	}

	works, _, err := getArtistWorks("nm0000000", post)
	assert.NoError(t, err)
	assert.Len(t, works, 2)
	assert.Equal(t, []string{"", "CURSOR1"}, cursors)
}

func TestGetArtistWorksGraphqlError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"errors":[{"message":"PersistedQueryNotFound"}]}`))
	}))
	defer ts.Close()

	post := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return http.Post(ts.URL, contentType, body)
	}

	_, _, err := getArtistWorks("nm0000000", post)
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

	_, _, err := getArtistWorks("nm9999999", post)
	assert.Error(t, err)
}
