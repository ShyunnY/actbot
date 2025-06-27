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
