package hackeronePrograms

import (
	"testing"

	testsuite "github.com/nbr23/atomic-banquet/utils"
)

func TestH1ProgramParse(t *testing.T) {
	testsuite.TestParseSuccess(
		t,
		HackeronePrograms{},
		map[string]interface{}{},
		1,
		`^\[[^]]+\] .* launched a program on [-\d:.TZ]+.*$`,
	)
}
