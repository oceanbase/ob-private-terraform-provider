package ocpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoRequestBasicAuthAndSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "admin" || p != "secret" {
			t.Errorf("期望 Basic Auth admin:secret，实际 %s:%s ok=%v", u, p, ok)
		}
		if r.URL.Path != "/api/v2/foo" {
			t.Errorf("路径不符 %s", r.URL.Path)
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
		t.Fatalf("意外错误：%v", err)
	}
	if out.Hello != "world" {
		t.Errorf("期望 world，实际 %q", out.Hello)
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
		t.Fatal("期望返回错误")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("期望 *NotFoundError，实际 %T：%v", err, err)
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
		t.Fatal("期望返回错误")
	}
	if got := err.Error(); got == "" || got[:8] != "OCP API " {
		t.Errorf("错误格式不符：%s", got)
	}
}
