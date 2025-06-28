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

package syncer

import (
	"io"
	"testing"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/stretchr/testify/assert"

	"github.com/ShyunnY/actbot/internal/actors"
)

func TestSyncerCommentBodyMatch(t *testing.T) {
	cases := []struct {
		caseName string
		comment  string
		expect   bool
	}{
		{
			caseName: "Match sync instruction",
			comment:  "/sync",
			expect:   true,
		},
		{
			caseName: "Match sync instruction with spaces",
			comment:  "/sync    ",
			expect:   true,
		},
		{
			caseName: "Unmatched instruction",
			comment:  "/resync",
			expect:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			assert.Equal(t, tc.expect, syncRegexp.MatchString(tc.comment))
		})
	}
}

func TestSyncerCapture(t *testing.T) {
	cases := []struct {
		caseName string
		event    actors.GenericEvent
		expect   bool
	}{
		{
			caseName: "Capture sync command",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/sync"),
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
						Body: github.Ptr[string]("/resync"),
					},
					Issue: &github.Issue{
						PullRequestLinks: nil,
					},
				},
			},
			expect: false,
		},
		{
			caseName: "Do not capture pull requests",
			event: actors.GenericEvent{
				Event: github.IssueCommentEvent{
					Comment: &github.IssueComment{
						Body: github.Ptr[string]("/sync"),
					},
					Issue: &github.Issue{
						PullRequestLinks: &github.PullRequestLinks{},
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			syncerActor := &actor{
				logger: slog.NewWithConfig(func(l *slog.Logger) {
					l.PushHandler(handler.NewIOWriterHandler(io.Discard, slog.AllLevels))
				}),
			}
			assert.Equal(t, tc.expect, syncerActor.Capture(tc.event))
		})
	}
}
