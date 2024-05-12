package hackeronePrograms

import (
	"testing"

	"github.com/nbr23/rss-banquet/parser"
	testsuite "github.com/nbr23/rss-banquet/utils"
)

func TestH1ProgramParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		HackeronePrograms{},
		&parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "results_count", Type: "int", Value: 10},
			},
			Parser: HackeronePrograms{},
		},
		1,
		`^\[[^]]+\] .* launched a program on [-\d:.TZ]+.*$`,
	)
}
