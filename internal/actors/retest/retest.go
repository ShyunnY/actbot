package retest

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
	"github.com/hashicorp/go-multierror"

	"github.com/ShyunnY/actbot/internal/actors"
)

const (
	retestActorName = "AssignActor"

	failedConclusion = "failure"
)

var retestRegexp = regexp.MustCompile(`^/retest\s*$`)

type actor struct {
	ghClient *github.Client
	logger   *slog.Logger

	event github.IssueCommentEvent
}

func NewRetestActor(ghClient *github.Client, logger *slog.Logger, _ *actors.Options) actors.Actor {
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
		loginUser       = comment.GetUser().GetLogin()
	)
	a.logger.Infof("actor %s started processing events, pr number: #%d", a.Name(), issue.GetNumber())

	pr, err := actors.GetPRFromIssue(a.ghClient, repo.GetFullName(), issue)
	if err != nil {
		return err
	}

	checkRuns, _, err := a.ghClient.Checks.ListCheckRunsForRef(
		context.Background(),
		owner,
		repoName,
		pr.GetHead().GetSHA(),
		nil,
	)
	if err != nil {
		return err
	}

	var failedRuns []*github.CheckRun
	if checkRuns.CheckRuns == nil {
		return nil
	}
	for _, run := range checkRuns.CheckRuns {
		if run.GetConclusion() == failedConclusion {
			failedRuns = append(failedRuns, run)
		}
	}

	if len(failedRuns) == 0 {
		if err := actors.AddComment(
			a.ghClient,
			fmt.Sprintf("@%s %s", loginUser, "The current checks run has all been run successfully and there is no need to rerun it again"),
			repo.GetFullName(),
			issue.GetNumber(),
		); err != nil {
			return err
		}
	} else {
		if err := actors.AddReaction(a.ghClient, actors.RocketReaction, repo.GetFullName(), comment.GetID()); err != nil {
			a.logger.Errorf("failed to add reaction %s to #%d comment in #%d issue", actors.RocketReaction, issue.GetNumber(), comment.GetID())
		}

		errG := multierror.Append(nil)
		for _, run := range failedRuns {
			if _, err := a.ghClient.Actions.RerunJobByID(
				context.Background(),
				owner,
				repoName,
				run.GetID(),
			); err != nil {
				a.logger.Errorf("failed to rerun failed '%s' job by err: %v", run.GetName(), err)
				errG = multierror.Append(errG, err)
				continue
			}
			a.logger.Infof("success to rerun failed '%s' job", run.GetName())
		}

		if errG.Unwrap() != nil {
			return errG.Unwrap()
		}
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

	if !retestRegexp.MatchString(commentEvent.Comment.GetBody()) {
		return false
	}
	a.event = commentEvent

	return true
}

func (a *actor) Name() string {
	return retestActorName
}
