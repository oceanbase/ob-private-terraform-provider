package ocpclient

import (
	"context"
	"errors"
	"fmt"
)

type CreateArbitrationParam struct {
	HostID            int64     `json:"hostId,omitempty"`
	RpmName           string    `json:"rpmName"`
	InstallPath       string    `json:"installPath,omitempty"`
	RunPath           string    `json:"runPath,omitempty"`
	RunUser           string    `json:"runUser,omitempty"`
	SvrPort           int       `json:"svrPort,omitempty"`
	Description       string    `json:"description,omitempty"`
	StartupParameters []KVParam `json:"startupParameters,omitempty"`
	ClientToken       string    `json:"clientToken,omitempty"`
}

type ArbitrationService struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type createArbResp struct {
	ID                   int64 `json:"id"`
	ArbitrationServiceID int64 `json:"arbitrationServiceId"`
}

// IsArbitrationSupported 探测当前 OCP 是否支持仲裁服务（仅企业版）
func (c *Client) IsArbitrationSupported(ctx context.Context) (bool, error) {
	err := c.doRequest(ctx, "GET", "/api/v2/arbitration/services", nil, nil)
	if err == nil {
		return true, nil
	}
	var nfe *NotFoundError
	if errors.As(err, &nfe) {
		return false, nil
	}
	return false, err
}

func (c *Client) CreateArbitration(ctx context.Context, param CreateArbitrationParam) (int64, int64, error) {
	var resp createArbResp
	if err := c.doRequest(ctx, "POST", "/api/v2/arbitration/services", param, &resp); err != nil {
		return 0, 0, err
	}
	if err := c.WaitForTask(ctx, resp.ID); err != nil {
		return resp.ID, resp.ArbitrationServiceID, err
	}
	return resp.ID, resp.ArbitrationServiceID, nil
}

func (c *Client) GetArbitration(ctx context.Context, id int64) (*ArbitrationService, error) {
	var a ArbitrationService
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v2/arbitration/services/%d", id), nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *Client) DeleteArbitration(ctx context.Context, id int64) error {
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v2/arbitration/services/%d/delete", id), nil, &resp); err != nil {
		return err
	}
	return c.WaitForTask(ctx, resp.ID)
}
