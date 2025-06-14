package assign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommentBodyMatch(t *testing.T) {
	cases := []struct {
		caseName string
		comment  string
		expect   [][]string
	}{
		{
			caseName: "Match the assign instruction",
			comment:  "/assign",
			expect: [][]string{
				{
					"/assign",
					"",
				},
			},
		},
		{
			caseName: "Match the unassign instruction",
			comment:  "/unassign",
			expect: [][]string{
				{
					"/unassign",
					"un",
				},
			},
		},
		{
			caseName: "unmatched instructions",
			comment:  "/foo",
			expect:   nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			match := assignRegexp.FindAllStringSubmatch(tc.comment, -1)
			if tc.expect != nil {
				assert.NotNil(t, match)
				assert.ElementsMatch(t, tc.expect, match)
			} else {
				assert.Nil(t, match)
			}
		})
	}
}
