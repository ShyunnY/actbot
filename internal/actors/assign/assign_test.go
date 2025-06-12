package assign

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	oauthGh "golang.org/x/oauth2/github"
)

func TestCommentBodyMatch(t *testing.T) {
	cases := []struct {
		caseName string
		comment  string
		expect   [][]string
	}{
		{
			caseName: "Match the assign instruction",
			comment:  "/assign",
			expect: [][]string{
				{
					"/assign",
					"",
				},
			},
		},
		{
			caseName: "Match the unassign instruction",
			comment:  "/unassign",
			expect: [][]string{
				{
					"/unassign",
					"un",
				},
			},
		},
		{
			caseName: "unmatched instructions",
			comment:  "/foo",
			expect:   nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			match := assignRegexp.FindAllStringSubmatch(tc.comment, -1)
			if tc.expect != nil {
				assert.NotNil(t, match)
				assert.ElementsMatch(t, tc.expect, match)
			} else {
				assert.Nil(t, match)
			}
		})
	}
}

func TestDemo(t *testing.T) {
	ghToken := os.Getenv("GITHUB_TOKEN")
	ghClient, err := initGitHubClient(ghToken)
	assert.NoError(t, err)

	issueComment, _, err := ghClient.Issues.GetComment(context.Background(), "ShyunnY", "sw-go", 2965380428)
	assert.NoError(t, err)

	issue, _, _ := ghClient.Issues.Get(context.Background(), "ShyunnY", "sw-go", 1)
	assert.NotNil(t, issue)

	slog.Println(*issueComment)
}

func initGitHubClient(ghToken string) (*github.Client, error) {
	if len(ghToken) == 0 {
		return nil, errors.New("empty github token")
	}

	oauthConfig := oauth2.Config{Endpoint: oauthGh.Endpoint}
	oClient := oauthConfig.Client(
		context.Background(),
		&oauth2.Token{AccessToken: ghToken},
	)

	ghClient := github.NewClient(oClient)

	return ghClient, nil
}
