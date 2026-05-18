package ocpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDoRequestBasicAuthAndSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "admin" || p != "secret" {
			t.Errorf("expected Basic Auth admin:secret, got %s:%s ok=%v", u, p, ok)
		}
		if r.URL.Path != "/api/v2/foo" {
			t.Errorf("path mismatch %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"successful": true,
			"status":     200,
			"data":       map[string]any{"hello": "world"},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "admin", "secret")
	var out struct {
		Hello string `json:"hello"`
	}
	if err := c.doRequest(context.Background(), "GET", "/api/v2/foo", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Hello != "world" {
		t.Errorf("expected world, got %q", out.Hello)
	}
}

func TestDoRequest404ReturnsNotFoundError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"successful":false,"status":404,"error":{"code":"NOT_FOUND","message":"not found"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "u", "p")
	err := c.doRequest(context.Background(), "GET", "/api/v2/x/1", nil, nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T: %v", err, err)
	}
}

func TestDoRequestOCPErrorIsFormatted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"successful":false,"status":400,"error":{"code":"COM10001","message":"bad name"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "u", "p")
	err := c.doRequest(context.Background(), "POST", "/api/v2/things", map[string]any{"x": 1}, nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := err.Error(); got == "" || got[:8] != "OCP API " {
		t.Errorf("error format mismatch: %s", got)
	}
}

func TestWaitForTaskSuccessful(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		status := "RUNNING"
		if calls >= 2 {
			status = "SUCCESSFUL"
		}
		_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":1,"status":"` + status + `","name":"x"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 10 * time.Millisecond
	c.PollingTimeout = 1 * time.Second
	if err := c.WaitForTask(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls < 2 {
		t.Errorf("expected at least 2 polls, got %d", calls)
	}
}

func TestWaitForTaskFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":7,"status":"FAILED","name":"x"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 5 * time.Millisecond
	err := c.WaitForTask(context.Background(), 7)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "task #7 failed") {
		t.Errorf("error message mismatch: %q", err.Error())
	}
}

func TestWaitForTaskFireAndForget(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "u", "p")
	c.TaskMode = TaskModeFireAndForget
	if err := c.WaitForTask(context.Background(), 99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("fire-and-forget mode should not issue HTTP requests")
	}
}
