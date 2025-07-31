package cc

import (
	"io"
	"testing"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/stretchr/testify/assert"

	"github.com/ShyunnY/actbot/internal/actors"
)

func TestCcCapture(t *testing.T) {
	cases := []struct {
		caseName string
		event    actors.GenericEvent
		expect   bool
		addCC    []string
		removeCC []string
	}{
		{
			caseName: "cc actor capture and handle reviewers add events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/cc @foo"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{
							URL: github.Ptr("https://github.com/example_owner/example_repo/pull/1234567890"),
						},
					},
				},
			},
			addCC:  []string{"foo"},
			expect: true,
		},
		{
			caseName: "cc actor capture and handle multi reviewers add events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/cc @foo @bar @baz"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{
							URL: github.Ptr("https://github.com/example_owner/example_repo/pull/1234567890"),
						},
					},
				},
			},
			addCC:  []string{"foo", "bar", "baz"},
			expect: true,
		},
		{
			caseName: "cc actor capture and handle reviewers remove events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/uncc @foo"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{
							URL: github.Ptr("https://github.com/example_owner/example_repo/pull/1234567890"),
						},
					},
				},
			},
			removeCC: []string{"foo"},
			expect:   true,
		},
		{
			caseName: "cc actor capture and handle multi reviewers remove events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/uncc @foo @bar @baz"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{
							URL: github.Ptr("https://github.com/example_owner/example_repo/pull/1234567890"),
						},
					},
				},
			},
			removeCC: []string{"foo", "bar", "baz"},
			expect:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			ccActor := &actor{
				// a noop logger for testing only
				logger: slog.NewWithConfig(func(l *slog.Logger) {
					l.PushHandler(handler.NewIOWriterHandler(io.Discard, slog.AllLevels))
				}),
			}

			assert.Equal(t, tc.expect, ccActor.Capture(tc.event))

			switch {
			case len(tc.addCC) > 0:
				assert.Equal(t, tc.addCC, ccActor.reviewers)
				assert.True(t, ccActor.cc)
			case len(tc.removeCC) > 0:
				assert.Equal(t, tc.removeCC, ccActor.reviewers)
				assert.False(t, ccActor.cc)
			}
		})
	}
}
