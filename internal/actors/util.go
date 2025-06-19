package actors

import (
	"context"
	"strings"

	"github.com/google/go-github/v72/github"
)

func AddComment(ghClient *github.Client, content, fullName string, issueNumber int) error {
	owner, repo := GetOwnerRepo(fullName)
	if _, _, err := ghClient.Issues.CreateComment(
		context.Background(),
		owner,
		repo,
		issueNumber,
		&github.IssueComment{
			Body: &content,
		},
	); err != nil {
		return err
	}

	return nil
}

func AddLabelToIssue(ghClient *github.Client, fullName string, issueNumber int, label ...string) error {
	owner, repo := GetOwnerRepo(fullName)
	if _, _, err := ghClient.Issues.AddLabelsToIssue(
		context.Background(),
		owner,
		repo,
		issueNumber,
		label,
	); err != nil {
		return err
	}

	return nil
}

func RemoveLabelToIssue(ghClient *github.Client, fullName string, issueNumber int, label string) error {
	owner, repo := GetOwnerRepo(fullName)

	issue, _, err := ghClient.Issues.Get(
		context.Background(),
		owner,
		repo,
		issueNumber,
	)
	switch {
	case err != nil:
		return err
	case issue == nil || len(issue.Labels) == 0:
		return nil
	default:
		ret := true
		for _, issueLabel := range issue.Labels {
			if issueLabel.GetName() == label {
				ret = false
			}
		}
		if ret {
			return nil
		}
	}

	if _, err := ghClient.Issues.RemoveLabelForIssue(
		context.Background(),
		owner,
		repo,
		issueNumber,
		label,
	); err != nil {
		return err
	}

	return nil
}

func AddReaction(ghClient *github.Client, reaction, fullName string, issueCommentID int64) error {
	owner, repo := GetOwnerRepo(fullName)
	if _, _, err := ghClient.Reactions.CreateIssueCommentReaction(
		context.Background(),
		owner,
		repo,
		issueCommentID,
		reaction,
	); err != nil {
		return err
	}

	return nil
}

func GetPRFromIssue(ghClient *github.Client, fullName string, issue *github.Issue) (*github.PullRequest, error) {
	owner, repo := GetOwnerRepo(fullName)
	pullRequest, _, err := ghClient.PullRequests.Get(
		context.Background(),
		owner,
		repo,
		issue.GetNumber(),
	)
	if err != nil {
		return nil, err
	}

	return pullRequest, nil
}

func GetOwnerRepo(fullName string) (owner, repo string) {
	split := strings.Split(fullName, "/")
	owner = split[0]
	repo = split[1]

	return
}
