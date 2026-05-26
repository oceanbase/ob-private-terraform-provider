package ocpclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateClusterReturnsIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/ob/clusters":
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			var p CreateClusterParam
			_ = json.Unmarshal(body, &p)
			if p.Name != "tf_test" || len(p.Zones) != 1 || p.Zones[0].RpmName == "" {
				t.Errorf("request body mismatch: %+v", p)
			}
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":30295,"clusterId":8,"status":"RUNNING"}}`))
		case "/api/v2/tasks/30295":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":30295,"status":"SUCCESSFUL"}}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	taskID, clusterID, err := c.CreateCluster(context.Background(), CreateClusterParam{
		Name: "tf_test", Type: "PRIMARY", Password: "x",
		Zones: []ZoneParam{{Name: "zone1", IdcName: "IDCA", Servers: []int64{1}, RpmName: "ob.rpm"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taskID != 30295 || clusterID != 8 {
		t.Errorf("result mismatch: taskID=%d clusterID=%d", taskID, clusterID)
	}
}

func TestGetClusterReturnsModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/ob/clusters/8" {
			t.Errorf("path mismatch %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":8,"name":"tf_test","status":"RUNNING","type":"PRIMARY"}}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	cl, err := c.GetCluster(context.Background(), 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.ID != 8 || cl.Name != "tf_test" {
		t.Errorf("response mismatch: %+v", cl)
	}
}

func TestDeleteClusterPolls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/ob/clusters/8":
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":40001}}`))
		case "/api/v2/tasks/40001":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":40001,"status":"SUCCESSFUL"}}`))
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	if err := c.DeleteCluster(context.Background(), 8); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListClusters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"contents":[{"id":1,"name":"a","status":"RUNNING"},{"id":2,"name":"b","status":"RUNNING"}]}}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	lst, err := c.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lst) != 2 {
		t.Errorf("expected 2 clusters, got %d", len(lst))
	}
}
