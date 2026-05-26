package ocpclient

import (
	"context"
	"fmt"
)

type CreateTenantParam struct {
	Name                 string                 `json:"name"`
	Mode                 string                 `json:"mode,omitempty"`
	PrimaryZone          string                 `json:"primaryZone,omitempty"`
	Charset              string                 `json:"charset,omitempty"`
	Collation            string                 `json:"collation,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Whitelist            string                 `json:"whitelist,omitempty"`
	TimeZone             string                 `json:"timeZone,omitempty"`
	RootPassword         string                 `json:"rootPassword"`
	EnableArbitration    bool                   `json:"enableArbitration,omitempty"`
	SkipImportTenantInfo bool                   `json:"skipImportTenantInfo,omitempty"`
	ServiceName          string                 `json:"serviceName,omitempty"`
	LoadType             string                 `json:"loadType,omitempty"`
	Zones                []TenantZoneParam      `json:"zones"`
	Parameters           []TenantParameterParam `json:"parameters"`
	ClientToken          string                 `json:"clientToken,omitempty"`
}

type TenantZoneParam struct {
	Name         string    `json:"name"`
	ReplicaType  string    `json:"replicaType"`
	ResourcePool PoolParam `json:"resourcePool"`
}

type PoolParam struct {
	UnitSpecName string `json:"unitSpecName"`
	UnitCount    int64  `json:"unitCount"`
}

type TenantParameterParam struct {
	Name          string `json:"name"`
	Value         string `json:"value"`
	ParameterType string `json:"parameterType,omitempty"`
}

type Tenant struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Mode   string `json:"mode"`
}

type createTenantResp struct {
	ID       int64 `json:"id"`
	TenantID int64 `json:"tenantId"`
}

func (c *Client) CreateTenant(ctx context.Context, clusterID int64, param CreateTenantParam) (int64, int64, error) {
	var resp createTenantResp
	path := fmt.Sprintf("/api/v2/ob/clusters/%d/tenants/createTenant", clusterID)
	if err := c.doRequest(ctx, "POST", path, param, &resp); err != nil {
		return 0, 0, err
	}
	if err := c.WaitForTask(ctx, resp.ID); err != nil {
		return resp.ID, resp.TenantID, err
	}
	return resp.ID, resp.TenantID, nil
}

func (c *Client) GetTenant(ctx context.Context, clusterID, tenantID int64) (*Tenant, error) {
	var t Tenant
	path := fmt.Sprintf("/api/v2/ob/clusters/%d/tenants/%d", clusterID, tenantID)
	if err := c.doRequest(ctx, "GET", path, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) DeleteTenant(ctx context.Context, clusterID, tenantID int64) error {
	var resp struct {
		ID int64 `json:"id"`
	}
	path := fmt.Sprintf("/api/v2/ob/clusters/%d/tenants/%d", clusterID, tenantID)
	if err := c.doRequest(ctx, "DELETE", path, nil, &resp); err != nil {
		return err
	}
	return c.WaitForTask(ctx, resp.ID)
}
