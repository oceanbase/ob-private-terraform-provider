package ocpclient

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type BatchCreateHostParam struct {
	HostBasicDataList []HostBasicData `json:"hostBasicDataList"`
	SshPort           int             `json:"sshPort"`
	Kind              string          `json:"kind"`
	IdcID             int64           `json:"idcId"`
	TypeID            int64           `json:"typeId"`
	CredentialID      int64           `json:"credentialId"`
	VpcID             *int64          `json:"vpcId,omitempty"`
	Alias             string          `json:"alias,omitempty"`
	Description       string          `json:"description,omitempty"`
}

type HostBasicData struct {
	InnerIpAddress string `json:"innerIpAddress"`
	SerialNumber   string `json:"serialNumber,omitempty"`
}

type Host struct {
	ID             int64  `json:"id"`
	InnerIpAddress string `json:"innerIpAddress"`
	Status         string `json:"status"`
	IdcID          int64  `json:"idcId"`
	IdcName        string `json:"idcName"`
	TypeID         int64  `json:"typeId"`
	TypeName       string `json:"typeName"`
	Alias          string `json:"alias"`
}

type batchCreateHostResp struct {
	ID      int64   `json:"id"`
	HostIDs []int64 `json:"hostIds"`
}

type HostListFilter struct {
	Status string
	IdcID  *int64
}

func (c *Client) BatchCreateHost(ctx context.Context, param BatchCreateHostParam) (int64, []int64, error) {
	var resp batchCreateHostResp
	if err := c.doRequest(ctx, "POST", "/api/v2/compute/hosts/batchCreate", param, &resp); err != nil {
		return 0, nil, err
	}
	if err := c.WaitForTask(ctx, resp.ID); err != nil {
		return resp.ID, resp.HostIDs, err
	}
	return resp.ID, resp.HostIDs, nil
}

func (c *Client) GetHost(ctx context.Context, id int64) (*Host, error) {
	var h Host
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v2/compute/hosts/%d", id), nil, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

func (c *Client) DeleteHost(ctx context.Context, id int64) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v2/compute/hosts/%d", id), nil, nil)
}

func (c *Client) ListHosts(ctx context.Context, filter *HostListFilter) ([]Host, error) {
	q := url.Values{}
	if filter != nil {
		if filter.Status != "" {
			q.Set("status", filter.Status)
		}
		if filter.IdcID != nil {
			q.Set("idcId", strconv.FormatInt(*filter.IdcID, 10))
		}
	}
	path := "/api/v2/compute/hosts"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var result struct {
		Contents []Host `json:"contents"`
	}
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Contents, nil
}
