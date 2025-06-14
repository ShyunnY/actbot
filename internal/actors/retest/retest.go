package retest

import (
	"github.com/ShyunnY/actbot/internal/actors"
	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
)

const (
	retestActorName = "AssignActor"

	retestComment = "/retest"
)

type actor struct {
	ghClient *github.Client
	logger   *slog.Logger

	event github.IssueCommentEvent
}

func NewRetestActor(ghClient *github.Client, logger *slog.Logger) actors.Actor {
	return &actor{
		ghClient: ghClient,
		logger:   logger,
	}
}

func (a *actor) Handler() error {
	return nil
}

func (a *actor) Capture(event actors.GenericEvent) bool {
	genericEvent := event.Event
	commentEvent, ok := genericEvent.(github.IssueCommentEvent)
	if !ok {
		a.logger.Error("cannot extract event to github.IssueCommentEvent, please check event type")
		return false
	}

	if !commentEvent.Issue.IsPullRequest() {
		return false
	}
	if commentEvent.Issue.GetClosedBy() != nil || !commentEvent.Issue.GetClosedAt().IsZero() {
		return false
	}
	if commentEvent.Comment.GetBody() != retestComment {
		return false
	}
	a.event = commentEvent

	return true
}

func (a *actor) Name() string {
	return retestActorName
}
