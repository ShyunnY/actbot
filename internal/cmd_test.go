package internal

import (
	"os"
	"testing"

	"github.com/ShyunnY/actbot/internal/actors"
	"github.com/stretchr/testify/assert"
)

func TestDemo(t *testing.T) {
	ghToken := os.Getenv("GITHUB_TOKEN")
	eventPath := "/Users/shyunny/project/golang/actbot/mock/events/event.json"

	gitHubClient, err := initGitHubClient(ghToken)
	assert.NoError(t, err)

	err = dispatch(string(IssueComment), eventPath, gitHubClient)
	assert.NoError(t, err)
}

func TestDemo2(t *testing.T) {
	ghToken := os.Getenv("GITHUB_TOKEN")
	//eventPath := "/Users/shyunny/project/golang/actbot/mock/events/event.json"

	gitHubClient, err := initGitHubClient(ghToken)
	assert.NoError(t, err)

	err = actors.RemoveLabelToIssue(gitHubClient, "ShyunnY/actbot", 2, "help wanted")
	assert.NoError(t, err)
}
