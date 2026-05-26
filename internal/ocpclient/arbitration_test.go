package ocpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateArbitration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/arbitration/services":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":90001,"arbitrationServiceId":5}}`))
		case "/api/v2/tasks/90001":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":90001,"status":"SUCCESSFUL"}}`))
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	_, aid, err := c.CreateArbitration(context.Background(), CreateArbitrationParam{
		RpmName: "ob.rpm", HostID: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if aid != 5 {
		t.Errorf("aid mismatch: %d", aid)
	}
}

func TestIsArbitrationSupportedTrue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"contents":[]}}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	ok, err := c.IsArbitrationSupported(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected supported=true")
	}
}

func TestIsArbitrationSupportedFalse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"successful":false,"status":404,"error":{"code":"NF","message":"x"}}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	ok, err := c.IsArbitrationSupported(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected supported=false")
	}
}

func TestDeleteArbitration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/arbitration/services/5/delete":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":100001}}`))
		case "/api/v2/tasks/100001":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":100001,"status":"SUCCESSFUL"}}`))
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	if err := c.DeleteArbitration(context.Background(), 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
