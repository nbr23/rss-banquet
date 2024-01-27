package parser

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
)

type Parser interface {
	Parse(map[string]any) (*feeds.Feed, error)
	Help() string
	Route(*gin.Engine) gin.IRoutes
}

func DefaultedGet[T any](m map[string]any, k string, d T) T {
	if v, ok := m[k]; ok {
		if _, ok := v.(T); ok {
			return v.(T)
		}
	}
	return d
}

func DefaultedGetSlice[S ~[]T, T any](m map[string]any, k string, d S) S {
	if v, ok := m[k]; ok {
		if reflect.TypeOf(v).Kind() == reflect.Slice {
			ret := make(S, 0)
			for _, v := range v.([]interface{}) {
				if _, ok := v.(T); ok {
					ret = append(ret, v.(T))
				}
			}
			return ret
		}
	}
	return d
}

func GetLatestDate(dates []time.Time) time.Time {
	latestDate := dates[0]
	for _, date := range dates {
		if date.After(latestDate) {
			latestDate = date
		}
	}
	return latestDate
}

func GetGuid(ss []string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(ss))))
}

func GetRemoteFileLastModified(url string) (time.Time, error) {
	resp, err := http.Head(url)
	if err != nil {
		return time.Time{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return time.Time{}, fmt.Errorf("unable to fetch the update file, status code: %d", resp.StatusCode)
	}

	lastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		return time.Time{}, err
	}

	return lastModified, nil
}

func ServeFeed(c *gin.Context, feed *feeds.Feed) {
	switch c.Query("feedFormat") {
	case "rss":
		rss, err := feed.ToRss()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		c.Data(200, "application/rss+xml", []byte(rss))
		return
	case "json":
		json, err := feed.ToJSON()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		c.Data(200, "application/json", []byte(json))
		return
	// case "atom":
	default:
		atom, err := feed.ToAtom()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		c.Data(200, "application/atom+xml", []byte(atom))
		return
	}
}
