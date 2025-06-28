// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/go-github/v72/github"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/jinzhu/copier"
	"golang.org/x/oauth2"
	oauthGh "golang.org/x/oauth2/github"

	"github.com/ShyunnY/actbot/internal/actors"
	"github.com/ShyunnY/actbot/internal/actors/syncer/im"
)

// initialize the global logger
var logger = func() *slog.Logger {
	return slog.NewWithConfig(func(inner *slog.Logger) {
		consoleHandler := handler.NewConsoleHandler(slog.AllLevels)
		inner.ChannelName = "actbot"
		inner.AddHandler(consoleHandler)
	})
}()

func Setup() error {
	var (
		ghToken          = os.Getenv("token")
		ghEvent          = os.Getenv("GITHUB_EVENT_NAME")
		ghEventPath      = os.Getenv("GITHUB_EVENT_PATH")
		dingTalkWebHooks = os.Getenv("DINGTALK_WEBHOOK_TOKENS")
	)

	gitHubClient, err := InitGitHubClient(ghToken)
	if err != nil {
		exit("failed to init GitHub client by err: %v", err)
	}

	dingTalkClient := im.NewDingTalkClient(dingTalkWebHooks)
	if dingTalkClient == nil {
		exit("failed to init DingTalk client, webhook URL is empty")
	}

	if err := dispatch(ghEvent, ghEventPath, gitHubClient, &actors.Options{
		IMClient: dingTalkClient,
	}); err != nil {
		exit("failed to dispatch event by err: %v", err)
	}

	return nil
}

func dispatch(ghEvent, ghEventPath string, ghClient *github.Client, opts *actors.Options) error {
	if len(ghEvent) == 0 {
		return errors.New("empty github event")
	}
	ghEventBytes, err := readGitHubEvent(ghEventPath)
	if err != nil {
		return err
	}

	switch ghEvent {
	case string(IssueComment):
		var (
			evt          github.IssueCommentEvent
			genericEvent actors.GenericEvent
		)
		if err := json.Unmarshal(ghEventBytes, &evt); err != nil {
			return fmt.Errorf("unmarshal '%s' github event: %w", IssueComment, err)
		}
		genericEvent.Event = evt

		for _, fn := range actorMap[IssueComment] {
			event, err := copyEvent(&genericEvent)
			if err != nil {
				return err
			}

			actor := fn(ghClient, logger, opts)
			if actor.Capture(*event) {
				if err = actor.Handler(); err != nil {
					exit("actor %s handle by err: %s", actor.Name(), err)
				}

				logger.Infof("actor %s successfully handle %s event", actor.Name(), IssueComment)
			}
		}

	default:
		return errors.New("unsupported github event")
	}

	return nil
}

func readGitHubEvent(ghEventPath string) ([]byte, error) {
	if len(ghEventPath) == 0 {
		return nil, errors.New("empty github event path")
	}

	_, err := os.Stat(ghEventPath)
	if err != nil {
		return nil, err
	}

	eventBytes, err := os.ReadFile(ghEventPath)
	if err != nil {
		return nil, err
	}

	return eventBytes, nil
}

func InitGitHubClient(ghToken string) (*github.Client, error) {
	if len(ghToken) == 0 {
		return nil, errors.New("empty github token")
	}

	oauthConfig := oauth2.Config{
		Endpoint: oauthGh.Endpoint,
	}
	oClient := oauthConfig.Client(
		context.Background(),
		&oauth2.Token{AccessToken: ghToken},
	)
	ghClient := github.NewClient(oClient)

	return ghClient, nil
}

func copyEvent(src *actors.GenericEvent) (*actors.GenericEvent, error) {
	var dst actors.GenericEvent

	err := copier.Copy(&dst, src)
	if err != nil {
		return nil, err
	}

	return &dst, nil
}

func exit(format string, err ...any) {
	// avoid losing the call stack information
	logger.CallerSkip += 1

	logger.Errorf(format, err...)
	os.Exit(1)
}
