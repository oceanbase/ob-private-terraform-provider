package ocpclient

import (
	"context"
	"fmt"
)

type CreateObproxyParam struct {
	Name                string               `json:"name"`
	Password            string               `json:"password,omitempty"`
	ProxyroPassword     string               `json:"proxyroPassword,omitempty"`
	WorkMode            string               `json:"workMode,omitempty"`
	InstallPath         string               `json:"installPath,omitempty"`
	ObproxyInstallParam *InstallObproxyParam `json:"obproxyInstallParam,omitempty"`
	ObLinks             []ObLinkParam        `json:"obLinks,omitempty"`
	StartupParameters   []KVParam            `json:"startupParameters,omitempty"`
	ClientToken         string               `json:"clientToken,omitempty"`
}

type InstallObproxyParam struct {
	HostIDs      []int64 `json:"hostIds,omitempty"`
	Version      string  `json:"version,omitempty"`
	SqlPort      *int    `json:"sqlPort,omitempty"`
	ExporterPort *int    `json:"exporterPort,omitempty"`
}

type ObLinkParam struct {
	ClusterName string `json:"clusterName"`
	ObClusterID int64  `json:"obClusterId,omitempty"`
}

type Obproxy struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type createObproxyResp struct {
	ID               int64 `json:"id"`
	ObproxyClusterID int64 `json:"obproxyClusterId"`
}

func (c *Client) CreateObproxy(ctx context.Context, param CreateObproxyParam) (int64, int64, error) {
	var resp createObproxyResp
	if err := c.doRequest(ctx, "POST", "/api/v2/obproxy/clusters", param, &resp); err != nil {
		return 0, 0, err
	}
	if err := c.WaitForTask(ctx, resp.ID); err != nil {
		return resp.ID, resp.ObproxyClusterID, err
	}
	return resp.ID, resp.ObproxyClusterID, nil
}

func (c *Client) GetObproxy(ctx context.Context, id int64) (*Obproxy, error) {
	var o Obproxy
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v2/obproxy/clusters/%d", id), nil, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (c *Client) DeleteObproxy(ctx context.Context, id int64) error {
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v2/obproxy/clusters/%d", id), nil, &resp); err != nil {
		return err
	}
	return c.WaitForTask(ctx, resp.ID)
}
