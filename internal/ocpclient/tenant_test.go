package ocpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTenantReturnsIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/ob/clusters/8/tenants/createTenant":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":50001,"tenantId":42}}`))
		case "/api/v2/tasks/50001":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":50001,"status":"SUCCESSFUL"}}`))
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	taskID, tenantID, err := c.CreateTenant(context.Background(), 8, CreateTenantParam{
		Name: "t1", Mode: "MYSQL", RootPassword: "x",
		Zones:      []TenantZoneParam{{Name: "zone1", ReplicaType: "FULL", ResourcePool: PoolParam{UnitSpecName: "S1", UnitCount: 1}}},
		Parameters: []TenantParameterParam{},
	})
	if err != nil {
		t.Fatalf("意外错误：%v", err)
	}
	if taskID != 50001 || tenantID != 42 {
		t.Errorf("结果不符：taskID=%d tenantID=%d", taskID, tenantID)
	}
}

func TestGetTenant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/ob/clusters/8/tenants/42" {
			t.Errorf("路径不符 %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":42,"name":"t1","status":"NORMAL","mode":"MYSQL"}}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	tn, err := c.GetTenant(context.Background(), 8, 42)
	if err != nil {
		t.Fatalf("意外错误：%v", err)
	}
	if tn.ID != 42 || tn.Name != "t1" {
		t.Errorf("响应不符：%+v", tn)
	}
}

func TestDeleteTenant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/ob/clusters/8/tenants/42":
			if r.Method != "DELETE" {
				t.Errorf("期望 DELETE，实际 %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":60001}}`))
		case "/api/v2/tasks/60001":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":60001,"status":"SUCCESSFUL"}}`))
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	if err := c.DeleteTenant(context.Background(), 8, 42); err != nil {
		t.Fatalf("意外错误：%v", err)
	}
}
