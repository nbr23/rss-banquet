package parser

import (
	"flag"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/stretchr/testify/assert"
)

type stubParser struct {
	name    string
	options OptionsList
	parse   func(*Options) (*feeds.Feed, error)
	got     *Options
}

func (p *stubParser) String() string {
	return p.name
}

func (p *stubParser) GetOptions() Options {
	return Options{
		OptionsList: Options{OptionsList: p.options}.GetOptionsCopy(),
		Parser:      p,
	}
}

func (p *stubParser) Parse(o *Options) (*feeds.Feed, error) {
	p.got = o
	if p.parse != nil {
		return p.parse(o)
	}
	return &feeds.Feed{
		Title: p.name,
		Link:  &feeds.Link{Href: "https://example.com"},
	}, nil
}

func TestOptionsList_Get(t *testing.T) {
	tests := []struct {
		name      string
		options   OptionsList
		key       string
		want      interface{}
		wantErr   bool
		isDefault bool
	}{
		{
			name: "String option",
			options: OptionsList{
				{Flag: "str", Type: "string", Value: "test"},
			},
			key:       "str",
			want:      "test",
			isDefault: false,
		},
		{
			name: "String option with pointer",
			options: OptionsList{
				{Flag: "str", Type: "string", Value: func() interface{} { s := "test"; return &s }()},
			},
			key:       "str",
			want:      "test",
			isDefault: false,
		},
		{
			name: "StringSlice option",
			options: OptionsList{
				{Flag: "slice", Type: "stringSlice", Value: "a,b,c"},
			},
			key:       "slice",
			want:      []string{"a", "b", "c"},
			isDefault: false,
		},
		{
			name: "StringSlice option with pointer",
			options: OptionsList{
				{Flag: "slice", Type: "stringSlice", Value: func() interface{} { s := "a,b,c"; return &s }()},
			},
			key:       "slice",
			want:      []string{"a", "b", "c"},
			isDefault: false,
		},
		{
			name: "Int option",
			options: OptionsList{
				{Flag: "num", Type: "int", Value: "42"},
			},
			key:       "num",
			want:      42,
			isDefault: false,
		},
		{
			name: "Int option with int value",
			options: OptionsList{
				{Flag: "num", Type: "int", Value: 42},
			},
			key:       "num",
			want:      42,
			isDefault: false,
		},
		{
			name: "Int option with pointer",
			options: OptionsList{
				{Flag: "num", Type: "int", Value: func() interface{} { i := 42; return &i }()},
			},
			key:       "num",
			want:      42,
			isDefault: false,
		},
		{
			name: "Bool option true",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: "true"},
			},
			key:       "flag",
			want:      true,
			isDefault: false,
		},
		{
			name: "Bool option false",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: "false"},
			},
			key:       "flag",
			want:      false,
			isDefault: false,
		},
		{
			name: "Bool option 1",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: "1"},
			},
			key:       "flag",
			want:      true,
			isDefault: false,
		},
		{
			name: "Bool option 0",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: "0"},
			},
			key:       "flag",
			want:      false,
			isDefault: false,
		},
		{
			name: "Bool option with pointer",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: func() interface{} { b := true; return &b }()},
			},
			key:       "flag",
			want:      true,
			isDefault: false,
		},
		{
			name: "Bad option type",
			options: OptionsList{
				{Flag: "default", Type: "unknown", Value: 123},
			},
			key:     "default",
			wantErr: true,
		},
		{
			name:      "Option not found",
			options:   OptionsList{},
			key:       "nonexistent",
			wantErr:   true,
			isDefault: false,
		},

		// Test cases for defaults
		{
			name: "String option with default",
			options: OptionsList{
				{Flag: "str", Type: "string", Default: "default"},
			},
			key:       "str",
			want:      "default",
			isDefault: true,
		},
		{
			name: "StringSlice option with default",
			options: OptionsList{
				{Flag: "slice", Type: "stringSlice", Default: "a,b,c"},
			},
			key:       "slice",
			want:      []string{"a", "b", "c"},
			isDefault: true,
		},
		{
			name: "Int option with default",
			options: OptionsList{
				{Flag: "num", Type: "int", Default: "42"},
			},
			key:       "num",
			want:      42,
			isDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, d, err := tt.options.Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("OptionsList.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if b, ok := got.(*bool); ok && b != nil {
				got = *b
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OptionsList.Get() = %v, want %v", got, tt.want)
			}
			if d != tt.isDefault {
				t.Errorf("OptionsList.Get() default = %v, want %v", d, tt.isDefault)
			}
		})
	}
}

func TestGetGuid(t *testing.T) {
	assert.Equal(t, "8df752f5498e7360c8e6406dc21001c616ca18de32efa56cdc998a924344394f", GetGuid([]string{"a", "b"}))
	assert.Equal(t, GetGuid([]string{"a", "b"}), GetGuid([]string{"a", "b"}))
	assert.NotEqual(t, GetGuid([]string{"a", "b"}), GetGuid([]string{"b", "a"}))
	assert.NotEqual(t, GetGuid([]string{"a", "b"}), GetGuid([]string{"a", "c"}))
	assert.Len(t, GetGuid(nil), 64)
}

func TestGetLatestDate(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	assert.Equal(t, d2, GetLatestDate([]time.Time{d1, d2, d3}))
	assert.Equal(t, d2, GetLatestDate([]time.Time{d2, d1, d3}))
	assert.Equal(t, d1, GetLatestDate([]time.Time{d1}))
	assert.Equal(t, d1, GetLatestDate([]time.Time{d1, d1}))
	assert.True(t, GetLatestDate(nil).IsZero())
	assert.True(t, GetLatestDate([]time.Time{}).IsZero())
}

func TestFeedToText(t *testing.T) {
	feed := &feeds.Feed{
		Title: "Feed Title",
		Link:  &feeds.Link{Href: "https://example.com"},
		Items: []*feeds.Item{
			{
				Title:       "First",
				Link:        &feeds.Link{Href: "https://example.com/1"},
				Description: "  line one\nline two  ",
			},
			{
				Title:       "Second",
				Link:        &feeds.Link{Href: "https://example.com/2"},
				Description: "desc",
			},
		},
	}

	expected := "# Feed Title | https://example.com\n" +
		"- First\n\thttps://example.com/1\n\tline one\n\tline two\n" +
		"- Second\n\thttps://example.com/2\n\tdesc"

	assert.Equal(t, expected, FeedToText(feed))
}

func TestFeedToTextNoItems(t *testing.T) {
	feed := &feeds.Feed{
		Title: "Feed Title",
		Link:  &feeds.Link{Href: "https://example.com"},
	}

	assert.Equal(t, "# Feed Title | https://example.com", FeedToText(feed))
}

func TestGetFileTypeFromUrl(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/img.png", "png"},
		{"https://example.com/img.jpg?width=200&h=300", "jpg"},
		{"https://example.com/img.gif#fragment", "gif"},
		{"https://example.com/dir.d/img.jpeg", "jpeg"},
		{"https://example.com/IMG.JPG", "jpg"},
		{"https://example.com/img", ""},
		{"https://example.com/dir.d/img", ""},
		{"http://localhost/img", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			assert.Equal(t, tt.want, GetFileTypeFromUrl(tt.url))
		})
	}
}

func TestIsImageType(t *testing.T) {
	for _, ext := range []string{"png", "jpg", "jpeg", "gif", "PNG", "JPG"} {
		assert.True(t, IsImageType(ext), ext)
	}
	for _, ext := range []string{"webp", "svg", "mp4", ""} {
		assert.False(t, IsImageType(ext), ext)
	}
}

func TestValidateURLHost(t *testing.T) {
	allowed := []string{"costco.com", "infocon.org"}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "Exact host", url: "https://costco.com/product.html"},
		{name: "Subdomain", url: "https://www.costco.com/product.html"},
		{name: "Deep subdomain", url: "https://a.b.costco.com/product.html"},
		{name: "Uppercase host", url: "https://WWW.COSTCO.COM/product.html"},
		{name: "Plain http", url: "http://costco.com/"},
		{name: "Second allowed host", url: "https://infocon.org/"},
		{name: "Lookalike suffix", url: "https://costco.com.evil.com/", wantErr: true},
		{name: "Lookalike prefix", url: "https://evilcostco.com.br/", wantErr: true},
		{name: "Userinfo trick", url: "https://costco.com@evil.com/", wantErr: true},
		{name: "Unrelated host", url: "https://example.com/", wantErr: true},
		{name: "Bad scheme", url: "ftp://costco.com/", wantErr: true},
		{name: "File scheme", url: "file:///etc/passwd", wantErr: true},
		{name: "No host", url: "/product.html", wantErr: true},
		{name: "Unparseable", url: "https://co stco.com/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURLHost(tt.url, allowed)
			if !tt.wantErr {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			assert.IsType(t, &NotFoundError{}, err)
		})
	}
}

func TestValidateURLHostNoAllowedHosts(t *testing.T) {
	assert.Error(t, ValidateURLHost("https://costco.com/", nil))
}

func TestOptionsGetBool(t *testing.T) {
	o := Options{
		OptionsList: OptionsList{
			{Flag: "on", Type: "bool", Value: "true"},
			{Flag: "off", Type: "bool", Value: "false"},
			{Flag: "str", Type: "string", Value: "true"},
		},
		Parser: &stubParser{name: "stub"},
	}

	assert.True(t, o.GetBool("on"))
	assert.False(t, o.GetBool("off"))
	assert.False(t, o.GetBool("str"))
	assert.False(t, o.GetBool("missing"))
}

func TestOptionsGetFallsBackToParserOptions(t *testing.T) {
	p := &stubParser{
		name: "stub",
		options: OptionsList{
			{Flag: "fromParser", Type: "string", Default: "parserDefault"},
		},
	}
	o := Options{
		OptionsList: OptionsList{
			{Flag: "fromOptions", Type: "string", Value: "set"},
		},
		Parser: p,
	}

	v, isDefault := o.GeWithDefaultFlag("fromOptions")
	assert.Equal(t, "set", v)
	assert.False(t, isDefault)

	v, isDefault = o.GeWithDefaultFlag("fromParser")
	assert.Equal(t, "parserDefault", v)
	assert.True(t, isDefault)

	v, isDefault = o.GeWithDefaultFlag("unknown")
	assert.Nil(t, v)
	assert.True(t, isDefault)
}

func TestOptionsSet(t *testing.T) {
	o := Options{
		OptionsList: OptionsList{
			{Flag: "a", Type: "string", Value: "a"},
			{Flag: "b", Type: "string", Value: "b"},
		},
		Parser: &stubParser{name: "stub"},
	}

	o.Set("a", "updated")
	o.Set("missing", "ignored")

	assert.Equal(t, "updated", o.Get("a"))
	assert.Equal(t, "b", o.Get("b"))
}

func TestGetOptionsCopyIsolatesOptions(t *testing.T) {
	original := Options{
		OptionsList: OptionsList{
			{Flag: "a", Type: "string", Value: "a", Required: true, IsPath: true, Help: "help", Default: "d", ShortFlag: "s"},
		},
	}

	c := original.GetOptionsCopy()
	assert.Len(t, c, 1)
	assert.Equal(t, *original.OptionsList[0], *c[0])
	assert.NotSame(t, original.OptionsList[0], c[0])

	c[0].Value = "mutated"
	assert.Equal(t, "a", original.OptionsList[0].Value)
}

func TestGetFullOptions(t *testing.T) {
	p := &stubParser{
		name: "stub",
		options: OptionsList{
			{Flag: "own", Type: "string", Default: "ownDefault"},
		},
	}

	o := GetFullOptions(p)

	assert.Len(t, o.OptionsList, 3)
	assert.Equal(t, "feedFormat", o.OptionsList[0].Flag)
	assert.Equal(t, "rss", o.OptionsList[0].Default)
	assert.False(t, o.OptionsList[0].IsStatic)
	assert.Equal(t, "route", o.OptionsList[1].Flag)
	assert.Equal(t, "stub", o.OptionsList[1].Default)
	assert.True(t, o.OptionsList[1].IsStatic)
	assert.Equal(t, "own", o.OptionsList[2].Flag)
	assert.Equal(t, "stub", o.Get("route"))
	assert.Equal(t, "ownDefault", o.Get("own"))
}

func TestGetHelp(t *testing.T) {
	o := Options{
		OptionsList: OptionsList{
			{Flag: "a", Type: "string", Default: "d", Help: "help a"},
			{Flag: "b", Type: "int", Default: "1", Help: "help b"},
		},
	}

	assert.Equal(t, "\t - a: help a (default: d)\n\t - b: help b (default: 1)\n", o.GetHelp())
}

func TestAddFlags(t *testing.T) {
	o := Options{
		OptionsList: OptionsList{
			{Flag: "str", Type: "string", Default: "strDefault"},
			{Flag: "slice", Type: "stringSlice", Default: "a,b"},
			{Flag: "num", Type: "int", Default: "3"},
			{Flag: "badNum", Type: "int", Default: "notAnInt"},
			{Flag: "flag", Type: "bool", Default: "true"},
		},
	}
	o.Parser = &stubParser{name: "stub", options: o.OptionsList}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	o.AddFlags(flags)

	assert.Equal(t, "strDefault", o.Get("str"))
	assert.Equal(t, []string{"a", "b"}, o.Get("slice"))
	assert.Equal(t, 3, o.Get("num"))
	assert.Equal(t, 0, o.Get("badNum"))
	assert.True(t, o.GetBool("flag"))

	err := flags.Parse([]string{"-str", "value", "-slice", "c,d", "-num", "7", "-flag=false"})
	assert.NoError(t, err)

	assert.Equal(t, "value", o.Get("str"))
	assert.Equal(t, []string{"c", "d"}, o.Get("slice"))
	assert.Equal(t, 7, o.Get("num"))
	assert.False(t, o.GetBool("flag"))
}

func TestAddFlagsUnknownType(t *testing.T) {
	o := Options{
		OptionsList: OptionsList{
			{Flag: "weird", Type: "unknown"},
		},
	}

	assert.Panics(t, func() {
		o.AddFlags(flag.NewFlagSet("test", flag.ContinueOnError))
	})
}

func TestSortFeedEntries(t *testing.T) {
	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	feed := &feeds.Feed{
		Items: []*feeds.Item{
			{Title: "older", Created: older},
			{Title: "newer", Created: newer},
		},
	}

	SortFeedEntries(feed)

	assert.Equal(t, "newer", feed.Items[0].Title)
	assert.Equal(t, "older", feed.Items[1].Title)

	empty := &feeds.Feed{}
	SortFeedEntries(empty)
	assert.Empty(t, empty.Items)
}

func TestErrorTypes(t *testing.T) {
	notFound := NewNotFoundError("missing thing")
	internal := NewInternalError("broken thing")

	assert.Equal(t, "NotFoundError: missing thing", notFound.Error())
	assert.Equal(t, "InternalError: broken thing", internal.Error())

	var err error = notFound
	_, ok := err.(*NotFoundError)
	assert.True(t, ok)

	err = internal
	_, ok = err.(*InternalError)
	assert.True(t, ok)
}
