package nytimes

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const NYT_GRAPHQL_HASH = "f9d85c0c99ec31bf5aac7d6e20a3f7beadbf7478877c56d4be4504aca19490c4"

func getNYTimesToken() (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://www.nytimes.com/"))
	if err != nil {
		return "", fmt.Errorf("failed to fetch token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch token, status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	body := string(bodyBytes)
	tokenStart := strings.Index(body, `"nyt-token":"`)
	if tokenStart == -1 {
		return "", fmt.Errorf("nyt-token not found in response")
	}
	tokenStart += len(`"nyt-token":"`)
	tokenEnd := strings.Index(body[tokenStart:], `"`)
	if tokenEnd == -1 {
		return "", fmt.Errorf("nyt-token end not found in response")
	}
	token := body[tokenStart : tokenStart+tokenEnd]
	return token, nil
}

func getGraphQLQuery(author string) (string, string) {
	extensions := fmt.Sprintf(`{"persistedQuery":{"version":1,"sha256Hash":"%s"}}`, NYT_GRAPHQL_HASH)
	variables := fmt.Sprintf(`{"id":"/by/%s","first":10}`, author)

	return url.QueryEscape(variables), url.QueryEscape(extensions)
}

func getGraphQLResponse(author string) ([]Edge, error) {
	token, err := getNYTimesToken()
	if err != nil {
		return nil, err
	}
	variables, extensions := getGraphQLQuery(author)
	myurl := fmt.Sprintf("https://samizdat-graphql.nytimes.com/graphql/v2?operationName=BylineQuery&variables=%s&extensions=%s", variables, extensions)

	req, err := http.NewRequest("GET", myurl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:138.0) Gecko/20100101 Firefox/138.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Referer", "https://www.nytimes.com/")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("nyt-app-type", "project-vi")
	req.Header.Set("nyt-app-version", "0.0.5")
	req.Header.Set("nyt-token", token)
	req.Header.Set("x-nyt-internal-meter-override", "undefined")
	req.Header.Set("Origin", "https://www.nytimes.com")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Priority", "u=4")
	req.Header.Set("TE", "trailers")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch articles, status code: %d", resp.StatusCode)
	}

	zipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	body, err := io.ReadAll(zipReader)

	if err != nil {
		return nil, err
	}
	var graphqlResponse GraphQLResponse
	err = json.Unmarshal(body, &graphqlResponse)
	if err != nil {
		return nil, err
	}
	return graphqlResponse.Data.AnyWork.ContentSearch.Hits.Edges, nil
}
