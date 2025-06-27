package internal

import (
	"github.com/ShyunnY/actbot/internal/actors/area"
	"github.com/ShyunnY/actbot/internal/actors/syncer"
	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"

	"github.com/ShyunnY/actbot/internal/actors"
	"github.com/ShyunnY/actbot/internal/actors/assign"
	"github.com/ShyunnY/actbot/internal/actors/retest"
)

type GitHubEventType string

type RegisterFn = func(ghClient *github.Client, logger *slog.Logger, opts *actors.Options) actors.Actor

const (
	IssueComment GitHubEventType = "issue_comment"
)

var actorMap = map[GitHubEventType][]RegisterFn{
	IssueComment: {
		assign.NewAssignActor,
		retest.NewRetestActor,
		area.NewLabelerActor,
		syncer.NewSyncerActor,
	},
}
