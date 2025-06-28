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

package labeler

import (
	"io"
	"testing"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/stretchr/testify/assert"

	"github.com/ShyunnY/actbot/internal/actors"
)

func TestLabelerCommentBodyMatch(t *testing.T) {
	cases := []struct {
		caseName string
		comment  string
		expect   bool
	}{
		{
			caseName: "Match area instruction",
			comment:  "/area label1",
			expect:   true,
		},
		{
			caseName: "Match unarea instruction",
			comment:  "/unarea label1",
			expect:   true,
		},
		{
			caseName: "Match kind instruction",
			comment:  "/kind label1",
			expect:   true,
		},
		{
			caseName: "Match unkind instruction",
			comment:  "/unkind label1",
			expect:   true,
		},
		{
			caseName: "Unmatched instruction",
			comment:  "/label",
			expect:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			var regexpMatch bool
			if areaRegexp.MatchString(tc.comment) || unareaRegexp.MatchString(tc.comment) ||
				kindRegexp.MatchString(tc.comment) || unkindRegexp.MatchString(tc.comment) {
				regexpMatch = true
			}
			assert.Equal(t, tc.expect, regexpMatch)
		})
	}
}

func TestLabelerCapture(t *testing.T) {
	cases := []struct {
		caseName string
		event    actors.GenericEvent
		expect   bool
	}{
		{
			caseName: "Capture area command",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/area label1"),
					},
					Issue: &github.Issue{
						PullRequestLinks: nil,
					},
				},
			},
			expect: true,
		},
		{
			caseName: "Capture unarea command",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/unarea label1"),
					},
					Issue: &github.Issue{
						PullRequestLinks: nil,
					},
				},
			},
			expect: true,
		},
		{
			caseName: "Capture kind command",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/kind label1"),
					},
					Issue: &github.Issue{
						PullRequestLinks: nil,
					},
				},
			},
			expect: true,
		},
		{
			caseName: "Capture unkind command",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/unkind label1"),
					},
					Issue: &github.Issue{
						PullRequestLinks: nil,
					},
				},
			},
			expect: true,
		},
		{
			caseName: "Do not capture empty comment",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string](""),
					},
					Issue: &github.Issue{
						PullRequestLinks: nil,
					},
				},
			},
			expect: false,
		},
		{
			caseName: "Do not capture unmatched command",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/label label1"),
					},
					Issue: &github.Issue{
						PullRequestLinks: nil,
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			labelerActor := &actor{
				logger: slog.NewWithConfig(func(l *slog.Logger) {
					l.PushHandler(handler.NewIOWriterHandler(io.Discard, slog.AllLevels))
				}),
			}
			assert.Equal(t, tc.expect, labelerActor.Capture(tc.event))
		})
	}
}
