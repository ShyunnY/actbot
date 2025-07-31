package cc

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"

	"github.com/ShyunnY/actbot/internal/actors"
)

const (
	ccActorName = "cc"
)

var ccRegexp = regexp.MustCompile(`(?mi)^/(un)?cc(\s+@[-/\w]+)+\s*$`)

type actor struct {
	ghClient *github.Client
	logger   *slog.Logger

	event github.IssueCommentEvent

	cc        bool
	reviewers []string
}

func NewCCActor(ghClient *github.Client, logger *slog.Logger) actors.Actor {
	return &actor{
		ghClient: ghClient,
		logger:   logger,
	}
}

func (a *actor) Handler() error {
	var (
		issue           = a.event.GetIssue()
		owner, repoName = actors.GetOwnerRepo(a.event.GetRepo().GetFullName())
	)
	a.logger.Infof("actor %s started processing events, issue number: #%d", a.Name(), issue.GetNumber())

	if a.cc {
		var statusCode int
		_, response, err := a.ghClient.PullRequests.RequestReviewers(
			context.Background(),
			owner,
			repoName,
			issue.GetNumber(),
			github.ReviewersRequest{
				Reviewers: a.reviewers,
			},
		)

		if response != nil {
			statusCode = response.StatusCode
		}
		if err != nil || statusCode == http.StatusUnprocessableEntity {
			return fmt.Errorf("failed to request reviewers for issue %d. Response status code: %d, err: %w", issue.GetNumber(), statusCode, err)
		}
		a.logger.Infof("actor %s requested reviewers for issue #%d. reviewers: [%s]", a.Name(), issue.GetNumber(), strings.Join(a.reviewers, ","))
	} else {
		_, err := a.ghClient.PullRequests.RemoveReviewers(
			context.Background(),
			owner,
			repoName,
			issue.GetNumber(),
			github.ReviewersRequest{
				Reviewers: a.reviewers,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to remove reviewers for issue %d. err: %w", issue.GetNumber(), err)
		}
		a.logger.Infof("actor %s remove reviewers for issue #%d. reviewers: [%s]", a.Name(), issue.GetNumber(), strings.Join(a.reviewers, ","))
	}

	return nil
}

func (a *actor) Capture(event actors.GenericEvent) bool {
	genericEvent := event.Event
	commentEvent, ok := genericEvent.(github.IssueCommentEvent)
	if !ok {
		a.logger.Error("cannot extract event to github.IssueCommentEvent, please check event type")
		return false
	}

	if !commentEvent.Issue.IsPullRequest() || len(commentEvent.Comment.GetBody()) == 0 {
		return false
	}
	if commentEvent.Issue.GetClosedBy() != nil || !commentEvent.Issue.GetClosedAt().IsZero() {
		return false
	}

	matches := ccRegexp.FindAllStringSubmatch(commentEvent.Comment.GetBody(), -1)
	if matches == nil {
		return false
	}

	var reviewers []string
	for _, match := range matches {
		// verify the command structure
		split := strings.Split(strings.TrimSpace(match[0]), " ")
		if len(split) < 2 || len(split[0]) == 0 {
			return false
		}

		action := split[0]
		if strings.HasPrefix(action, "/un") {
			a.cc = false
		} else {
			a.cc = true
		}

		// get the reviewers for one or more applications
		for _, reviewer := range split[1:] {
			user := strings.TrimPrefix(strings.TrimSpace(reviewer), "@")
			if len(user) == 0 {
				continue
			}
			reviewers = append(reviewers, user)
		}
	}
	if len(reviewers) == 0 {
		a.logger.Infof("actor %s has no reviewers to Assignment", a.Name())
		return false
	}

	a.event = commentEvent
	a.reviewers = reviewers

	return true
}

func (a *actor) Name() string {
	return ccActorName
}
