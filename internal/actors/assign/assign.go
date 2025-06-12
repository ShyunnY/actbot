package assign

import (
	"encoding/json"
	"regexp"

	"github.com/ShyunnY/actbot/internal/actors"
	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
)

const (
	assignActorName = "AssignActor"
)

var (
	assignRegexp = regexp.MustCompile(`^/(un)?assign\b`)
)

type actor struct {
	ghClient *github.Client
	logger   *slog.Logger

	event github.IssueCommentEvent
	add   bool
}

func NewAssignActor(ghClient *github.Client, logger *slog.Logger) actors.Actor {
	return &actor{
		ghClient: ghClient,
		logger:   logger,
		add:      true,
	}
}

func (a *actor) Handler() error {
	a.logger.Infof("actor %s started processing events", a.Name())
	a.logger.Info(a.event)
	output, err := json.Marshal(a.event)
	if err != nil {
		a.logger.Error(err)
	}
	a.logger.Info(string(output))

	// 注意，issue 可以分配给多个人，并且通过 assignees 字段来确定（在 add 中就需要进一步确认了）

	// 当然还需要忽略 pr，因为 pr 也是一个 issue（这个可以直接判断出来）
	// 根据 add 来决定执行：
	// 需要根据当前登陆的用户来执行操作
	//	*add：查看当前的 issue 是否已经被分配了（还有！是否是分配给自己了），如果被分配了则需要写入一个 comment。如果存在 help wanted label，还需要移除
	//	*remove：查看当前的 issue 是否已经被分配了，如果没分配了则需要写入一个 comment。如果不存在 help wanted label，还需要新增
	var (
		issue     = a.event.GetIssue()
		comment   = a.event.GetComment()
		loginUser = comment.GetUser()
		repo      = a.event.GetRepo()
		assignees = issue.Assignees
	)

	if a.add {
		if isAssignLoginUser(loginUser, assignees) {
			// todo: 不允许分配, 机器人需要写一个评论并 @ 用户
		}

		// 执行分配
		owner := repo.Owner.Name
		repoName := repo.Name

		a.logger.Info("repo = %s, owner = %s", repoName, owner)
		//_, _, err := a.ghClient.Issues.AddAssignees(context.Background(), *owner, *repoName, *issue.Number, []string{loginUser.GetLogin()})
		//if err != nil {
		//	return err
		//}

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

	// pull request is essentially an issue, and the current actor does not handle this situation.
	if commentEvent.Issue.IsPullRequest() {
		return false
	}
	comment := commentEvent.GetComment()
	if comment == nil || len(comment.GetBody()) == 0 {
		return false
	}

	matches := assignRegexp.FindAllStringSubmatch(comment.GetBody(), -1)
	if matches == nil {
		return false
	}
	for _, match := range matches {
		if match[1] == "un" {
			a.add = false
		}
	}

	return true
}

func (a *actor) Name() string {
	return assignActorName
}

func isAssignLoginUser(user *github.User, assignees []*github.User) bool {
	if assignees == nil || len(assignees) == 0 {
		return false
	}

	for _, assignee := range assignees {
		if assignee.GetID() == user.GetID() {
			return true
		}
	}

	return false
}
