package ocpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateObproxy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/obproxy/clusters":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":70001,"obproxyClusterId":11}}`))
		case "/api/v2/tasks/70001":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":70001,"status":"SUCCESSFUL"}}`))
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	_, obID, err := c.CreateObproxy(context.Background(), CreateObproxyParam{
		Name:                "px1",
		ObproxyInstallParam: &InstallObproxyParam{HostIDs: []int64{2}, Version: "obproxy.rpm"},
		ObLinks:             []ObLinkParam{{ClusterName: "cl1"}},
	})
	if err != nil {
		t.Fatalf("意外错误：%v", err)
	}
	if obID != 11 {
		t.Errorf("obID 不符：%d", obID)
	}
}

func TestGetObproxy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/obproxy/clusters/11" {
			t.Errorf("路径不符 %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":11,"name":"px1","status":"RUNNING"}}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	o, err := c.GetObproxy(context.Background(), 11)
	if err != nil {
		t.Fatalf("意外错误：%v", err)
	}
	if o.Name != "px1" {
		t.Errorf("响应不符：%+v", o)
	}
}

func TestDeleteObproxy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/obproxy/clusters/11":
			if r.Method != "DELETE" {
				t.Errorf("期望 DELETE，实际 %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":80001}}`))
		case "/api/v2/tasks/80001":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":80001,"status":"SUCCESSFUL"}}`))
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	if err := c.DeleteObproxy(context.Background(), 11); err != nil {
		t.Fatalf("意外错误：%v", err)
	}
}
