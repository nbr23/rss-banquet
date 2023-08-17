package parser

import (
	"github.com/gorilla/feeds"
)

type Parser interface {
	Parse(map[string]any) (*feeds.Feed, error)
	Help() string
}

func DefaultedGet(m map[string]any, k string, d any) any {
	if v, ok := m[k].(string); ok {
		return v
	}
	return d
}
