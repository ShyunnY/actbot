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

package im

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type DingTalkClient struct {
	WebhookURL string
}

func NewDingTalkClient(webhookURL string) *DingTalkClient {
	return &DingTalkClient{WebhookURL: webhookURL}
}

func (d *DingTalkClient) SendMessage(issueNumber int64, content string) error {

	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	message := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title":   "Issue #" + strconv.Itoa(int(issueNumber)) + " Sync, please pay annotation",
			"content": content,
		},
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", d.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}
