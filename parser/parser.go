package parser

import (
	"time"

	"github.com/gorilla/feeds"
)

type Parser interface {
	Parse(map[string]any) (*feeds.Feed, error)
	Help() string
}

func DefaultedGet[T any](m map[string]any, k string, d T) T {
	if v, ok := m[k]; ok {
		if _, ok := v.(T); ok {
			return v.(T)
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
