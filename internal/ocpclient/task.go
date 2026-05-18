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

// WaitForTask polls until the task reaches SUCCESSFUL, FAILED, or timeout.
// In TaskModeFireAndForget mode, it returns nil immediately.
func (c *Client) WaitForTask(ctx context.Context, taskID int64) error {
	if c.TaskMode == TaskModeFireAndForget {
		return nil
	}

	deadline := time.Now().Add(c.PollingTimeout)
	for time.Now().Before(deadline) {
		task, err := c.getTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to poll task #%d: %w", taskID, err)
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
