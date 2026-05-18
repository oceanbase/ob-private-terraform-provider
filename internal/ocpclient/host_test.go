package ocpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatchCreateHostReturnsIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/compute/hosts/batchCreate":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":110001,"hostIds":[201,202]}}`))
		case "/api/v2/tasks/110001":
			_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"id":110001,"status":"SUCCESSFUL"}}`))
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	c.PollingInterval = 1
	_, ids, err := c.BatchCreateHost(context.Background(), BatchCreateHostParam{
		HostBasicDataList: []HostBasicData{{InnerIpAddress: "1.1.1.1"}, {InnerIpAddress: "1.1.1.2"}},
		SshPort:           22, Kind: "DEDICATED_PHYSICAL_MACHINE", IdcID: 1, TypeID: 1, CredentialID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != 201 {
		t.Errorf("ids mismatch: %v", ids)
	}
}

func TestGetHostNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"successful":false}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	_, err := c.GetHost(context.Background(), 999)
	if err == nil {
		t.Fatal("expected an error")
	}
	var nfe *NotFoundError
	if !asNotFoundError(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func asNotFoundError(err error, target **NotFoundError) bool {
	if e, ok := err.(*NotFoundError); ok {
		*target = e
		return true
	}
	return false
}

func TestListHosts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"successful":true,"status":200,"data":{"contents":[{"id":1,"innerIpAddress":"1.1.1.1","status":"AVAILABLE","idcId":1,"typeId":1}]}}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "u", "p")
	lst, err := c.ListHosts(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lst) != 1 {
		t.Errorf("expected 1 host, got %d", len(lst))
	}
}
