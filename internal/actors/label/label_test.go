package label

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

func TestLabelCapture(t *testing.T) {
	cases := []struct {
		caseName     string
		event        actors.GenericEvent
		expect       bool
		addLabels    []string
		removeLabels []string
	}{
		{
			caseName: "label actor capture and handle label add events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/label kind/chore"),
					},
					Issue: &github.Issue{},
				},
			},
			expect:    true,
			addLabels: []string{"kind/chore"},
		},
		{
			caseName: "label actor capture and handle space split label add events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/label help wanted"),
					},
					Issue: &github.Issue{},
				},
			},
			expect:    true,
			addLabels: []string{"help wanted"},
		},
		{
			caseName: "label actor capture and handle multi line space split label add events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string](`
						/label help wanted
						/label kind/chore
						`,
						),
					},
					Issue: &github.Issue{},
				},
			},
			expect:    true,
			addLabels: []string{"help wanted", "kind/chore"},
		},
		{
			caseName: "label actor capture and handle label remove events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/unlabel kind/chore"),
					},
					Issue: &github.Issue{},
				},
			},
			expect:       true,
			removeLabels: []string{"kind/chore"},
		},
		{
			caseName: "label actor capture and handle space split label remove events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/unlabel help wanted"),
					},
					Issue: &github.Issue{},
				},
			},
			expect:       true,
			removeLabels: []string{"help wanted"},
		},
		{
			caseName: "label actor capture and handle multi line space split label remove events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string](
							`/unlabel help wanted
							/unlabel kind/chore`,
						),
					},
					Issue: &github.Issue{},
				},
			},
			expect:       true,
			removeLabels: []string{"help wanted", "kind/chore"},
		},
		{
			caseName: "label actor capture and handle hybrid label events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string](
							`
							/label area/first
							/label area/second		
							/unlabel help wanted
							/unlabel kind/chore`,
						),
					},
					Issue: &github.Issue{},
				},
			},
			expect:       true,
			addLabels:    []string{"area/first", "area/second"},
			removeLabels: []string{"help wanted", "kind/chore"},
		},
		{
			caseName: "label actor does not capture pull request",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/label good first issue"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{},
					},
				},
			},
			expect: false,
		},
		{
			caseName: "label actor does not capture closed issue",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/label good first issue"),
					},
					Issue: &github.Issue{
						ClosedAt: &github.Timestamp{Time: time.Now()},
					},
				},
			},
			expect: false,
		},
		{
			caseName: "label actor does not capture empty comment body issue",
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
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			labelActor := &actor{
				// a noop logger for testing only
				logger: slog.NewWithConfig(func(l *slog.Logger) {
					l.PushHandler(handler.NewIOWriterHandler(io.Discard, slog.AllLevels))
				}),
			}
			assert.Equal(t, tc.expect, labelActor.Capture(tc.event))
			assert.Equal(t, tc.addLabels, labelActor.addLabels)
			assert.Equal(t, tc.removeLabels, labelActor.removeLabels)
		})
	}
}
