package area

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
	labelPrefix      = "area/"
)

var (
	areaRegexp   = regexp.MustCompile(`^/area\s+(.+)$`)
	unareaRegexp = regexp.MustCompile(`^/unarea\s+(.+)$`)
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
		labels := strings.Fields(areaMatch[1]) // 分割多个标签
		for _, label := range labels {
			label = labelPrefix + label
			err = a.checkAndAddLabel(owner, repoName, issue.GetNumber(), label)
			if err != nil {
				return err // 如果添加失败，返回错误
			}
		}
	} else if unareaMatch := unareaRegexp.FindStringSubmatch(body); unareaMatch != nil {
		labels := strings.Fields(unareaMatch[1]) // 分割多个标签
		for _, label := range labels {
			label = labelPrefix + label
			err = actors.RemoveLabelToIssue(a.ghClient, repo.GetFullName(), issue.GetNumber(), label)
			if err != nil {
				return err // 如果移除失败，返回错误
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

	if areaRegexp.MatchString(commentEvent.Comment.GetBody()) || unareaRegexp.MatchString(commentEvent.Comment.GetBody()) {
		a.event = commentEvent
		return true
	}

	return false
}

func (a *actor) Name() string {
	return labelerActorName
}
