package label

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/ShyunnY/actbot/internal/actors"
)

const (
	labelActorName = "LabelActor"

	labelPrefix   = "/label"
	unlabelPrefix = "/unlabel"
)

var (
	labelRegex   = regexp.MustCompile(`(?m)^\s*/label\s*(.*?)\s*$`)
	unlabelRegex = regexp.MustCompile(`(?m)^\s*/unlabel\s*(.*?)\s*$`)
)

type actor struct {
	ghClient *github.Client
	logger   *slog.Logger

	addLabels    []string
	removeLabels []string

	event github.IssueCommentEvent
}

func NewLabelActor(ghClient *github.Client, logger *slog.Logger) actors.Actor {
	return &actor{
		ghClient: ghClient,
		logger:   logger,
	}
}

func (a *actor) Handler() error {
	var (
		issue           = a.event.GetIssue()
		repo            = a.event.GetRepo()
		comment         = a.event.GetComment()
		loginUser       = comment.GetUser()
		owner, repoName = actors.GetOwnerRepo(repo.GetFullName())
	)
	a.logger.Infof("actor %s started processing events, pr number: #%d", a.Name(), issue.GetNumber())

	repoLabels := sets.Set[string]{}
	listLabels, _, err := a.ghClient.Issues.ListLabels(
		context.Background(),
		owner,
		repoName,
		nil,
	)
	if err != nil {
		return err
	}
	for _, listLabel := range listLabels {
		if len(listLabel.GetName()) != 0 {
			repoLabels.Insert(listLabel.GetName())
		}
	}

	issueLabels := sets.Set[string]{}
	for _, label := range issue.Labels {
		if len(label.GetName()) != 0 {
			issueLabels.Insert(label.GetName())
		}
	}

	var nonExistRepoLabels []string
	for _, addLabel := range a.addLabels {
		if issueLabels.Has(addLabel) {
			continue
		}

		// label does not exist in the current repo, we need to record the event
		if !repoLabels.Has(addLabel) {
			nonExistRepoLabels = append(nonExistRepoLabels, addLabel)
			continue
		}

		if err := actors.AddLabelToIssue(a.ghClient, repo.GetFullName(), issue.GetNumber(), addLabel); err != nil {
			a.logger.Errorf("actor %s failed to add '%s' label to actor %s: %v", a.Name(), labelActorName, addLabel, err)
			return err
		}
	}

	var nonExistIssueLabels []string
	for _, removeLabel := range a.removeLabels {
		// current issue does not have a label, we need to record the event
		if !issueLabels.Has(removeLabel) {
			nonExistIssueLabels = append(nonExistIssueLabels, removeLabel)
			continue
		}

		if !repoLabels.Has(removeLabel) {
			continue
		}

		if err := actors.RemoveLabelToIssue(a.ghClient, repo.GetFullName(), issue.GetNumber(), removeLabel); err != nil {
			a.logger.Errorf("actor %s failed to remove '%s' label to actor %s: %v", a.Name(), labelActorName, removeLabel, err)
			return err
		}
	}

	switch {
	case len(nonExistRepoLabels) > 0:
		a.logger.Warnf("The repo is missing labels '(%s)'", strings.Join(nonExistRepoLabels, ","))
		returnMsg := fmt.Sprintf("These labels '(%s)' cannot be used because they are not configured in the repo.", strings.Join(nonExistRepoLabels, ","))
		return actors.AddComment(a.ghClient, fmt.Sprintf("@%s %s", loginUser.GetLogin(), returnMsg), repo.GetFullName(), issue.GetNumber())

	case len(nonExistIssueLabels) > 0:
		returnMsg := fmt.Sprintf("These labels '(%s)' cannot be applied to issues because they are not exist in the issue.", strings.Join(nonExistIssueLabels, ","))
		return actors.AddComment(a.ghClient, fmt.Sprintf("@%s %s", loginUser.GetLogin(), returnMsg), repo.GetFullName(), issue.GetNumber())
	default:
		return nil
	}
}

func (a *actor) Capture(event actors.GenericEvent) bool {
	genericEvent := event.Event
	commentEvent, ok := genericEvent.(github.IssueCommentEvent)
	if !ok {
		a.logger.Error("cannot extract event to github.IssueCommentEvent, please check event type")
		return false
	}

	// pull request is essentially an issue, and the current actor does not handle this situation.
	if commentEvent.Issue.IsPullRequest() {
		return false
	}

	// do not handle closed issues
	if !commentEvent.Issue.GetClosedAt().IsZero() || commentEvent.Issue.ClosedBy != nil {
		return false
	}

	comment := commentEvent.GetComment()
	if comment == nil || len(comment.GetBody()) == 0 {
		return false
	}

	labelMatchers := labelRegex.FindAllStringSubmatch(comment.GetBody(), -1)
	unlabelMatchers := unlabelRegex.FindAllStringSubmatch(comment.GetBody(), -1)
	if len(labelMatchers) == 0 && len(unlabelMatchers) == 0 {
		return false
	}

	// refine label matching rules
	validateMatchFn := func(matches [][]string, prefix string) []string {
		var ret []string
		for _, match := range matches {
			matchParts := strings.Split(
				strings.TrimSpace(match[0]),
				" ",
			)
			if len(matchParts) < 2 || matchParts[0] != prefix {
				continue
			}

			// reason for using match instead of matchParts is that some labels may contain spaces.
			// e.g. "help wanted", "good first issue"...
			ret = append(ret, match[1])
		}
		return ret
	}

	addLabels := validateMatchFn(labelMatchers, labelPrefix)
	removeLabels := validateMatchFn(unlabelMatchers, unlabelPrefix)
	if len(addLabels) == 0 && len(removeLabels) == 0 {
		return false
	}

	a.addLabels = addLabels
	a.removeLabels = removeLabels
	a.event = commentEvent

	return true
}

func (a *actor) Name() string {
	return labelActorName
}
