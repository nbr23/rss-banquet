package parser

import (
	"crypto/sha256"
	"fmt"
	"reflect"
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
