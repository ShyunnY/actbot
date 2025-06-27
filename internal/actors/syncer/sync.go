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
	"fmt"
	"regexp"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"

	"github.com/ShyunnY/actbot/internal/actors"
)

const (
	syncerActorName = "SyncerActor"
)

var syncRegexp = regexp.MustCompile(`^/sync\s*$`)

type actor struct {
	ghClient *github.Client
	logger   *slog.Logger
	imClient actors.IMClient
	event    github.IssueCommentEvent
}

func NewSyncerActor(ghClient *github.Client, logger *slog.Logger, opts *actors.Options) actors.Actor {

	return &actor{
		ghClient: ghClient,
		logger:   logger,
		imClient: opts.IMClient,
	}
}

func (a *actor) Handler() error {

	issue := a.event.GetIssue()
	comment := a.event.GetComment()
	loginUser := comment.GetUser().GetLogin()

	message := fmt.Sprintf("User @%s has requested to sync issue #%d: %s", loginUser, issue.GetNumber(), issue.GetTitle())

	if err := a.imClient.SendMessage(*issue.ID, message); err != nil {
		a.logger.Errorf("failed to send message to IM: %v", err)
		return err
	}

	a.logger.Infof("message sent to IM: %s", message)
	return nil
}

func (a *actor) Capture(event actors.GenericEvent) bool {

	genericEvent := event.Event
	commentEvent, ok := genericEvent.(github.IssueCommentEvent)
	if !ok {
		a.logger.Error("cannot extract event to github.IssueCommentEvent, please check event type")
		return false
	}

	if commentEvent.Issue.IsPullRequest() || len(commentEvent.Comment.GetBody()) == 0 {
		return false
	}
	if !syncRegexp.MatchString(commentEvent.Comment.GetBody()) {
		return false
	}

	a.event = commentEvent
	return true
}

func (a *actor) Name() string {
	return syncerActorName
}
