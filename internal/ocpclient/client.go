package ocpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TaskMode string

const (
	TaskModePolling       TaskMode = "polling"
	TaskModeFireAndForget TaskMode = "fire_and_forget"
)

type Client struct {
	BaseURL         string
	HTTPClient      *http.Client
	Username        string
	Password        string
	TaskMode        TaskMode
	PollingTimeout  time.Duration
	PollingInterval time.Duration
}

type NotFoundError struct{ Path string }

func (e *NotFoundError) Error() string { return "OCP 资源不存在：" + e.Path }

func NewClient(baseURL, username, password string) *Client {
	return &Client{
		BaseURL:         baseURL,
		Username:        username,
		Password:        password,
		HTTPClient:      &http.Client{Timeout: 60 * time.Second},
		TaskMode:        TaskModePolling,
		PollingTimeout:  30 * time.Minute,
		PollingInterval: 10 * time.Second,
	}
}

type ocpResponse struct {
	Data       json.RawMessage `json:"data"`
	Status     int             `json:"status"`
	Successful bool            `json:"successful"`
	Error      *ocpError       `json:"error,omitempty"`
}

type ocpError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败：%w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("构造请求失败：%w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("执行请求失败：%w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败：%w", err)
	}

	if resp.StatusCode == 404 {
		return &NotFoundError{Path: path}
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("OCP server error [HTTP %d]: %s", resp.StatusCode, string(respBody))
	}

	var ocpResp ocpResponse
	if err := json.Unmarshal(respBody, &ocpResp); err != nil {
		return fmt.Errorf("解析响应失败：%w（响应体：%s）", err, string(respBody))
	}
	if !ocpResp.Successful {
		if ocpResp.Error != nil {
			return fmt.Errorf("OCP API error [HTTP %d]: %s - %s", resp.StatusCode, ocpResp.Error.Code, ocpResp.Error.Message)
		}
		return fmt.Errorf("OCP API error [HTTP %d]: %s", resp.StatusCode, string(respBody))
	}
	if result != nil && len(ocpResp.Data) > 0 {
		if err := json.Unmarshal(ocpResp.Data, result); err != nil {
			return fmt.Errorf("解析响应 data 字段失败：%w", err)
		}
	}
	return nil
}
