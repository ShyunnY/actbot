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
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ShyunnY/actbot/internal/actors"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
)

const (
	labelerActorName = "LabelerActor"
	areaPrefix       = "area/"
	kindPrefix       = "kind/"
)

var (
	areaRegexp   = regexp.MustCompile(`^/area\s+(.+)$`)
	unareaRegexp = regexp.MustCompile(`^/unarea\s+(.+)$`)
	kindRegexp   = regexp.MustCompile(`^/kind\s+(.+)$`)
	unkindRegexp = regexp.MustCompile(`^/unkind\s+(.+)$`)
)

type actor struct {
	ghClient *github.Client
	logger   *slog.Logger

	event github.IssueCommentEvent
}

func NewLabelerActor(ghClient *github.Client, logger *slog.Logger, _ *actors.Options) actors.Actor {
	return &actor{
		ghClient: ghClient,
		logger:   logger,
	}
}

func (a *actor) Handler() error {

	var (
		issue           = a.event.GetIssue()
		repo            = a.event.GetRepo()
		owner, repoName = actors.GetOwnerRepo(repo.GetFullName())
		comment         = a.event.GetComment()
		body            = comment.GetBody()
	)

	a.logger.Infof("actor %s started processing events, issue number: #%d", a.Name(), issue.GetNumber())

	var err error
	if areaMatch := areaRegexp.FindStringSubmatch(body); areaMatch != nil {
		labels := strings.Fields(areaMatch[1]) // Split multiple labels
		for _, label := range labels {
			label = areaPrefix + label
			err = a.checkAndAddLabel(owner, repoName, issue.GetNumber(), label)
			if err != nil {
				return err // Return error if adding fails
			}
		}
	} else if unareaMatch := unareaRegexp.FindStringSubmatch(body); unareaMatch != nil {
		labels := strings.Fields(unareaMatch[1]) // Split multiple labels
		for _, label := range labels {
			label = areaPrefix + label
			err = actors.RemoveLabelToIssue(a.ghClient, repo.GetFullName(), issue.GetNumber(), label)
			if err != nil {
				return err // Return error if removing fails
			}
		}
	} else if kindMatch := kindRegexp.FindStringSubmatch(body); kindMatch != nil {
		labels := strings.Fields(kindMatch[1])
		for _, label := range labels {
			label = kindPrefix + label
			err = a.checkAndAddLabel(owner, repoName, issue.GetNumber(), label)
			if err != nil {
				return err
			}
		}
	} else if unkindMatch := unkindRegexp.FindStringSubmatch(body); unkindMatch != nil {
		labels := strings.Fields(unkindMatch[1])
		for _, label := range labels {
			label = kindPrefix + label
			err = actors.RemoveLabelToIssue(a.ghClient, repo.GetFullName(), issue.GetNumber(), label)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *actor) checkAndAddLabel(owner, repoName string, issueNumber int, label string) error {
	labels, _, err := a.ghClient.Issues.ListLabels(context.Background(), owner, repoName, nil)
	if err != nil {
		return err
	}

	labelExists := false
	for _, l := range labels {
		if l.GetName() == label {
			labelExists = true
			break
		}
	}

	if !labelExists {
		if err := actors.AddComment(
			a.ghClient,
			fmt.Sprintf("The label '%s' does not exist. Please create it first.", label),
			repoName,
			issueNumber,
		); err != nil {
			return err
		}
		return fmt.Errorf("label '%s' does not exist", label)
	}

	err = actors.AddLabelToIssue(a.ghClient, repoName, issueNumber, label)
	if err != nil {
		a.logger.Errorf("failed to add label '%s' to issue #%d: %v", label, issueNumber, err)
		return err
	}
	a.logger.Infof("added label '%s' to issue #%d", label, issueNumber)
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
	if commentEvent.Issue.GetClosedBy() != nil || !commentEvent.Issue.GetClosedAt().IsZero() {
		return false
	}

	if areaRegexp.MatchString(commentEvent.Comment.GetBody()) ||
		unareaRegexp.MatchString(commentEvent.Comment.GetBody()) ||
		kindRegexp.MatchString(commentEvent.Comment.GetBody()) ||
		unkindRegexp.MatchString(commentEvent.Comment.GetBody()) {
		a.event = commentEvent
		return true
	}

	return false
}

func (a *actor) Name() string {
	return labelerActorName
}
