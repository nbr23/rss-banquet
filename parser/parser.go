package parser

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
)

type Parser interface {
	Parse(*Options) (*feeds.Feed, error)
	GetOptions() Options
	String() string
}

func GetFullOptions(p Parser) *Options {
	opts := p.GetOptions()
	opts.OptionsList = append([]*Option{
		{
			Flag:     "feedFormat",
			Required: false,
			Type:     "string",
			Help:     "feed output format (rss, atom, json)",
			Default:  "rss",
			Value:    "",
		},
		{
			Flag:     "private",
			Required: false,
			Type:     "bool",
			Help:     "private feed",
			Default:  "false",
			Value:    false,
		},
		{
			Flag:     "route",
			Required: false,
			Type:     "string",
			Help:     "route to expose the feed",
			Default:  p.String(),
			Value:    p.String(),
			IsStatic: true,
		},
	}, opts.OptionsList...)

	return &opts
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
	case "json":
		json, err := feed.ToJSON()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		c.Data(200, "application/json", []byte(json))
		return
	case "atom":
		atom, err := feed.ToAtom()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		c.Data(200, "application/atom+xml", []byte(atom))
		return
	// case "rss":
	default:
		rss, err := feed.ToRss()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		c.Data(200, "application/rss+xml", []byte(rss))
		return
	}
}

type Option struct {
	Flag      string
	Value     interface{}
	Required  bool
	Default   string
	Help      string
	ShortFlag string
	Type      string
	IsPath    bool
	IsStatic  bool // static options are exposed through the API
}

type Options struct {
	OptionsList OptionsList
	Parser      Parser
}

type OptionsList []*Option

func (o OptionsList) Get(key string) (interface{}, error) {
	for _, option := range o {
		if option.Flag == key {
			switch option.Type {
			case "string":
				if str, ok := option.Value.(string); ok {
					return str, nil
				}
				return *(option.Value.(*string)), nil
			case "stringSlice":
				if str, ok := option.Value.(string); ok {
					return strings.Split(str, ","), nil
				}
				return strings.Split(*(option.Value.(*string)), ","), nil
			case "int":
				if str, ok := option.Value.(string); ok {
					i, err := strconv.Atoi(str)
					if err != nil {
						return 0, nil
					}
					return i, nil
				}
				if stri, ok := option.Value.(int); ok {
					return stri, nil
				}
				return *(option.Value.(*int)), nil
			case "bool":
				if option.Value == nil {
					return false, nil
				}
				if strp, ok := option.Value.(*string); ok {
					return *strp == "true" || *strp == "1", nil
				}
				if str, ok := option.Value.(string); ok {
					return str == "true" || str == "1", nil
				}
				return (option.Value.(bool)), nil
			default:
				return option.Value, nil
			}
		}
	}
	return nil, fmt.Errorf("option not found")
}

func (o *Options) Get(key string) interface{} {
	v, err := o.OptionsList.Get(key)
	if err == nil {
		return v
	}
	v, err = o.Parser.GetOptions().OptionsList.Get(key)
	if err == nil {
		return v
	}
	return nil
}

func (o Options) GetHelp() string {
	help := ""
	for _, option := range o.OptionsList {
		help += fmt.Sprintf("\t - %s: %s (default: %s)\n", option.Flag, option.Help, option.Default)
	}
	return help
}

func (o *Options) ParseYaml(m map[string]any) error {
	for _, option := range o.OptionsList {
		if v, ok := m[option.Flag]; ok {
			option.Value = v
		} else {
			option.Value = option.Default
		}
	}
	return nil
}

func Route(g *gin.Engine, p Parser, o *Options) gin.IRoutes {
	urlPath := []string{o.Get("route").(string)}
	for _, option := range o.OptionsList {
		if option.IsStatic {
			continue
		}
		if option.Required {
			prefix := ":"
			if option.IsPath {
				prefix = "*"
			}
			urlPath = append(urlPath, fmt.Sprintf("%s%s", prefix, option.Flag))
		}
	}
	route := fmt.Sprintf("/%s", strings.Join(urlPath, "/"))

	return g.GET(route, func(c *gin.Context) {
		for _, option := range o.OptionsList {
			if option.Required {
				if c.Param(option.Flag) == "" {
					c.String(400, "missing required parameter: %s", option.Flag)
					return
				} else {
					option.Value = c.Param(option.Flag)
				}
			} else {
				if c.Query(option.Flag) == "" {
					option.Value = option.Default
				} else {
					option.Value = c.Query(option.Flag)
				}
			}
		}
		feed, err := p.Parse(o)
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		ServeFeed(c, feed)
	})
}

func (o Options) AddFlags(f *flag.FlagSet) {
	for _, option := range o.OptionsList {
		switch option.Type {
		case "bool":
			option.Value = f.Bool(option.Flag, option.Default == "true", option.Help)
		case "int":
			d, err := strconv.Atoi(option.Default)
			if err != nil {
				d = 0
			}
			option.Value = f.Int(option.Flag, d, option.Help)
		case "string":
			option.Value = f.String(option.Flag, option.Default, option.Help)
		case "stringSlice":
			option.Value = f.String(option.Flag, option.Default, option.Help)
		default:
			panic(fmt.Errorf("unknown type: %s", option.Type))
		}
	}
}
