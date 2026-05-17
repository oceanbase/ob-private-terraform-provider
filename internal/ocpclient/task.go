package ocpclient

import (
	"context"
	"fmt"
	"time"
)

type taskDetail struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	Name   string `json:"name"`
}

func (c *Client) getTask(ctx context.Context, taskID int64) (*taskDetail, error) {
	var task taskDetail
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v2/tasks/%d", taskID), nil, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// WaitForTask 轮询直到任务 SUCCESSFUL、FAILED 或超时
// TaskModeFireAndForget 模式下立即返回 nil
func (c *Client) WaitForTask(ctx context.Context, taskID int64) error {
	if c.TaskMode == TaskModeFireAndForget {
		return nil
	}

	deadline := time.Now().Add(c.PollingTimeout)
	for time.Now().Before(deadline) {
		task, err := c.getTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("轮询任务 #%d 失败：%w", taskID, err)
		}
		switch task.Status {
		case "SUCCESSFUL":
			return nil
		case "FAILED":
			return fmt.Errorf("OCP task #%d failed. Please check details at %s/#/task/%d", taskID, c.BaseURL, taskID)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.PollingInterval):
		}
	}
	return fmt.Errorf("timed out waiting for OCP task #%d after %s. The task may still be running", taskID, c.PollingTimeout)
}
