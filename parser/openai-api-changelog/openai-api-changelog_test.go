package openaiapichangelog

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/testsuite"
)

const fixture = `<html><body>
<section>
<h3 class="_ChangelogSectionTitle_test">November, 2026</h3>
<div class="grid">
<div data-variant="outline">November 17th</div>
<div data-variant="soft">API</div>
<div data-variant="soft">Models</div>
<div class="_ChangelogMarkdown_test"><p>First changelog entry adds an API feature.</p></div>
</div>
<div class="grid">
<div data-variant="outline">November 16th</div>
<div data-variant="soft">API</div>
<div class="_ChangelogMarkdown_test"><p>Second changelog entry improves an API feature.</p></div>
</div>
<div class="grid">
<div data-variant="outline">November 15th</div>
<div data-variant="soft">API</div>
<div class="_ChangelogMarkdown_test"><p>Third changelog entry fixes an API feature.</p></div>
</div>
<div class="grid">
<div data-variant="outline">November 14th</div>
<div data-variant="soft">API</div>
<div class="_ChangelogMarkdown_test"><p>Fourth changelog entry updates an API feature.</p></div>
</div>
<div class="grid">
<div data-variant="outline">November 13th</div>
<div data-variant="soft">API</div>
<div class="_ChangelogMarkdown_test"><p>Fifth changelog entry removes an API feature.</p></div>
</div>
</section>
</body></html>`

func TestOpenAIAPIChangelogParse(t *testing.T) {
	previousGet := getChangelog
	getChangelog = func(gotURL string, options map[string]any) (*http.Response, error) {
		if gotURL != url {
			t.Fatalf("got URL %q, want %q", gotURL, url)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(fixture)),
		}, nil
	}
	t.Cleanup(func() { getChangelog = previousGet })

	testsuite.TestParseSuccess(
		t,
		OpenAIAPIChangelog{},
		&parser.Options{
			OptionsList: parser.OptionsList{},
			Parser:      OpenAIAPIChangelog{},
		},
		5,
		`.+`,
		`^OpenAI API Changelog$`,
	)
}
