package assign

import (
	"io"
	"testing"
	"time"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/stretchr/testify/assert"

	"github.com/ShyunnY/actbot/internal/actors"
)

func TestAssignCommentBodyMatch(t *testing.T) {
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

func TestAssignCapture(t *testing.T) {
	cases := []struct {
		caseName string
		event    actors.GenericEvent
		expect   bool
	}{
		{
			caseName: "assign actor capture and handle events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/assign"),
					},
					Issue: &github.Issue{},
				},
			},
			expect: true,
		},
		{
			caseName: "assign actor does not capture pull request",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/assign"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{},
					},
				},
			},
			expect: false,
		},
		{
			caseName: "assign actor does not capture closed issue",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/assign"),
					},
					Issue: &github.Issue{
						ClosedAt: &github.Timestamp{Time: time.Now()},
					},
				},
			},
			expect: false,
		},
		{
			caseName: "assign actor does not capture empty comment body issue",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string](""),
					},
					Issue: &github.Issue{},
				},
			},
			expect: false,
		},
		{
			caseName: "assign actor does not capture unmatched assignRegexp comment body issue",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/foo_assign"),
					},
					Issue: &github.Issue{},
				},
			},
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			assignActor := &actor{
				// a noop logger for testing only
				logger: slog.NewWithConfig(func(l *slog.Logger) {
					l.PushHandler(handler.NewIOWriterHandler(io.Discard, slog.AllLevels))
				}),
			}
			assert.Equal(t, tc.expect, assignActor.Capture(tc.event))
		})
	}
}
