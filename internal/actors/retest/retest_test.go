// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package retest

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

func TestRetestCommentBodyMatch(t *testing.T) {
	cases := []struct {
		caseName string
		comment  string
		expect   bool
	}{
		{
			caseName: "Match the retest instruction",
			comment:  "/retest",
			expect:   true,
		},
		{
			caseName: "Match the instructions that show multiple spaces after retest",
			comment:  "/retest    ",
			expect:   true,
		},
		{
			caseName: "unmatched instructions",
			comment:  "/redo",
			expect:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			assert.Equal(t, tc.expect, retestRegexp.MatchString(tc.comment))
		})
	}
}

func TestRetestCapture(t *testing.T) {
	cases := []struct {
		caseName string
		event    actors.GenericEvent
		expect   bool
	}{
		{
			caseName: "retest actor capture and handle events",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/retest"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{
							URL: github.Ptr("https://github.com/example_owner/example_repo/pull/1234567890"),
						},
					},
				},
			},
			expect: true,
		},
		{
			caseName: "retest actor does not capture issue that are not pull request",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/retest"),
					},
					Issue: &github.Issue{},
				},
			},
			expect: false,
		},
		{
			caseName: "retest actor does not capture empty comment pull request",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string](""),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{
							URL: github.Ptr("https://github.com/example_owner/example_repo/pull/1234567890"),
						},
					},
				},
			},
			expect: false,
		},
		{
			caseName: "retest actor does not capture closed pull request",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/retest"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{
							URL: github.Ptr("https://github.com/example_owner/example_repo/pull/1234567890"),
						},
						ClosedAt: &github.Timestamp{Time: time.Now()},
					},
				},
			},
			expect: false,
		},
		{
			caseName: "retest actor does not capture unmatched retestRegexp comment body pull request",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/retest1"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{
							URL: github.Ptr("https://github.com/example_owner/example_repo/pull/1234567890"),
						},
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			retestActor := &actor{
				// a noop logger for testing only
				logger: slog.NewWithConfig(func(l *slog.Logger) {
					l.PushHandler(handler.NewIOWriterHandler(io.Discard, slog.AllLevels))
				}),
			}
			assert.Equal(t, tc.expect, retestActor.Capture(tc.event))
		})
	}
}
