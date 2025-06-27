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
