package style

import (
	"strings"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/stretchr/testify/assert"
)

func TestInjectRssStyle(t *testing.T) {
	out := InjectRssStyle(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"></rss>`)

	assert.Contains(t, out, `<?xml-stylesheet type="text/xsl" href="/rss-style.xsl"?>`)
	assert.Contains(t, out, `<?xml version="1.0" encoding="utf-8"?>`)
	assert.NotContains(t, out, `encoding="UTF-8"`)
	assert.Contains(t, out, `<rss version="2.0">`)
}

func TestInjectAtomStyle(t *testing.T) {
	out := InjectAtomStyle(`<?xml version="1.0" encoding="UTF-8"?><feed></feed>`)

	assert.Contains(t, out, `<?xml-stylesheet type="text/xsl" href="/atom-style.xsl"?>`)
	assert.Contains(t, out, `<?xml version="1.0" encoding="utf-8"?>`)
	assert.Contains(t, out, `<feed>`)
}

func TestInjectStyleNoDeclaration(t *testing.T) {
	in := `<rss version="2.0"></rss>`

	assert.Equal(t, in, InjectRssStyle(in))
	assert.Equal(t, in, InjectAtomStyle(in))
}

func TestInjectStyleOnlyFirstOccurrence(t *testing.T) {
	in := `<?xml version="1.0" encoding="UTF-8"?><a><?xml version="1.0" encoding="UTF-8"?></a>`
	out := InjectRssStyle(in)

	assert.Equal(t, 1, strings.Count(out, `/rss-style.xsl`))
	assert.Contains(t, out, `<a><?xml version="1.0" encoding="UTF-8"?></a>`)
}

func TestInjectStyleOnGeneratedFeed(t *testing.T) {
	feed := &feeds.Feed{
		Title:   "Test Feed",
		Link:    &feeds.Link{Href: "https://example.com"},
		Created: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Items: []*feeds.Item{
			{
				Title:   "Item",
				Link:    &feeds.Link{Href: "https://example.com/item"},
				Created: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	rss, err := feed.ToRss()
	assert.NoError(t, err)
	assert.Contains(t, InjectRssStyle(rss), `<?xml-stylesheet type="text/xsl" href="/rss-style.xsl"?>`)

	atom, err := feed.ToAtom()
	assert.NoError(t, err)
	assert.Contains(t, InjectAtomStyle(atom), `<?xml-stylesheet type="text/xsl" href="/atom-style.xsl"?>`)
}
