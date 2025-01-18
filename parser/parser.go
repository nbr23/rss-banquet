package parser

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/rs/zerolog/log"

	"github.com/nbr23/rss-banquet/config"
	"github.com/nbr23/rss-banquet/style"
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
		},
		{
			Flag:     "route",
			Required: false,
			Type:     "string",
			Help:     "route to expose the feed",
			Default:  p.String(),
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

func FeedToText(f *feeds.Feed) string {
	var txt string
	txt += fmt.Sprintf("# %s | %s\n", f.Title, f.Link.Href)
	for _, i := range f.Items {
		txt += fmt.Sprintf("- %s\n", i.Title)
		txt += fmt.Sprintf("\t%s\n", i.Link.Href)
		txt += fmt.Sprintf("\t%s\n", strings.TrimSpace(strings.ReplaceAll(i.Description, "\n", "\n\t")))
	}
	return strings.TrimSpace(txt)
}

func ServeFeed(c *gin.Context, f *feeds.Feed) {
	switch c.Query("feedFormat") {
	case "json":
		json, err := f.ToJSON()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		c.Data(200, "application/json", []byte(json))
		return
	case "atom":
		atom, err := f.ToAtom()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		atom = style.InjectAtomStyle(atom)
		c.Data(200, "application/xml", []byte(atom))
		return
	case "text":
		text := FeedToText(f)
		c.Data(200, "text/plain", []byte(text))
		return
	// case "rss":
	default:
		rss, err := f.ToRss()
		if err != nil {
			c.String(500, "error parsing feed")
			return
		}
		rss = style.InjectRssStyle(rss)
		c.Data(200, "application/xml", []byte(rss))
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

func (o Options) GetOptionsCopy() OptionsList {
	var opts []*Option
	for _, option := range o.OptionsList {
		opts = append(opts, &Option{
			Flag:      option.Flag,
			Value:     option.Value,
			Required:  option.Required,
			Default:   option.Default,
			Help:      option.Help,
			ShortFlag: option.ShortFlag,
			Type:      option.Type,
			IsPath:    option.IsPath,
			IsStatic:  option.IsStatic,
		})
	}
	return opts
}

type Options struct {
	OptionsList OptionsList
	Parser      Parser
}

type OptionsList []*Option

func (o OptionsList) Get(key string) (interface{}, bool, error) {
	for _, option := range o {
		if option.Flag == key {
			switch option.Type {
			case "string":
				if str, ok := option.Value.(string); ok {
					return str, false, nil
				}
				if str, ok := option.Value.(*string); ok {
					return *str, false, nil
				}
				return option.Default, true, nil
			case "stringSlice":
				if str, ok := option.Value.(string); ok {
					return strings.Split(str, ","), false, nil
				}
				if str, ok := option.Value.(*string); ok {
					return strings.Split(*str, ","), false, nil
				}
				return strings.Split(option.Default, ","), true, nil
			case "int":
				if str, ok := option.Value.(string); ok {
					i, err := strconv.Atoi(str)
					if err == nil {
						return i, false, nil
					}
				}
				if stri, ok := option.Value.(int); ok {
					return stri, false, nil
				}
				if stri, ok := option.Value.(*int); ok {
					return *stri, false, nil
				}
				i, err := strconv.Atoi(option.Default)
				if err != nil {
					log.Error().Msgf("error parsing default value for option %s: %s", key, err)
					return 0, true, err
				}
				return i, true, nil
			case "bool":
				var b bool
				if strp, ok := option.Value.(*string); ok {
					b = *strp == "true" || *strp == "1"
					return &b, false, nil
				}
				if str, ok := option.Value.(string); ok {
					b = str == "true" || str == "1"
					return &b, false, nil
				}
				if b, ok := option.Value.(bool); ok {
					return &b, false, nil
				}
				if b, ok := option.Value.(*bool); ok {
					return b, false, nil
				}
				b = option.Default == "true" || option.Default == "1"
				return &b, true, nil
			default:
				return nil, true, fmt.Errorf("unknown type: %s", option.Type)
			}
		}
	}
	return nil, true, fmt.Errorf("option not found")
}

func (o *Options) Get(key string) interface{} {
	v, _ := o.GeWithDefaultFlag(key)
	return v
}

func (o *Options) GeWithDefaultFlag(key string) (interface{}, bool) {
	v, isDefault, err := o.OptionsList.Get(key)
	if err == nil {
		return v, isDefault
	}
	v, isDefault, err = o.Parser.GetOptions().OptionsList.Get(key)
	if err == nil {
		return v, isDefault
	}
	return nil, true
}

func (o *Options) Set(key string, value string) {
	for _, option := range o.OptionsList {
		if option.Flag == key {
			option.Value = value
		}
	}
}

func (o Options) GetHelp() string {
	help := ""
	for _, option := range o.OptionsList {
		help += fmt.Sprintf("\t - %s: %s (default: %s)\n", option.Flag, option.Help, option.Default)
	}
	return help
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
	route := fmt.Sprintf("/feed/%s", strings.Join(urlPath, "/"))

	return g.GET(route, func(c *gin.Context) {
		options := o.GetOptionsCopy()
		for _, option := range options {
			if option.Required {
				if c.Param(option.Flag) == "" {
					log.Error().Msgf("missing required parameter: %s", option.Flag)
					c.String(400, "missing required parameter: %s", option.Flag)
					return
				} else {
					option.Value = c.Param(option.Flag)
				}
			} else {
				if c.Query(option.Flag) != "" {
					option.Value = c.Query(option.Flag)
				}
			}
		}
		feed, err := p.Parse(&Options{OptionsList: options, Parser: p})
		if err != nil {
			switch err.(type) {
			case *NotFoundError:
				c.String(404, err.Error())
				return
			case *InternalError:
				c.String(500, err.Error())
				return
			default:
				log.Error().Msgf("error parsing feed: %s", err)
				c.String(500, "error parsing feed")
				return
			}
		}
		SortFeedEntries(feed)
		ServeFeed(c, feed)
	})
}

type NotFoundError struct {
	message string
}

type InternalError struct {
	message string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("NotFoundError: %s", e.message)
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("InternalError: %s", e.message)
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{message: message}
}

func NewInternalError(message string) *InternalError {
	return &InternalError{message: message}
}

func SortFeedEntries(f *feeds.Feed) {
	sort.Slice(f.Items, func(i, j int) bool {
		return f.Items[i].Created.After(f.Items[j].Created)
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

func GetFileTypeFromUrl(url string) string {
	parts := strings.Split(strings.Split(url, "?")[0], ".")

	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func IsImageType(t string) bool {
	switch t {
	case "png", "jpg", "jpeg", "gif":
		return true
	default:
		return false
	}
}

func HttpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	userAgent := config.GetConfigOption("USER_AGENT")
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
