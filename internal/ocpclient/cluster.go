package ocpclient

import (
	"context"
	"fmt"
)

type CreateClusterParam struct {
	Name                 string              `json:"name"`
	Type                 string              `json:"type"`
	Password             string              `json:"password,omitempty"`
	FullVersion          string              `json:"fullVersion,omitempty"`
	ObClusterID          *int64              `json:"obClusterId,omitempty"`
	Zones                []ZoneParam         `json:"zones"`
	Attributes           *ClusterAttributes  `json:"attributes,omitempty"`
	StartupParameters    []KVParam           `json:"startupParameters,omitempty"`
	PrimaryZone          string              `json:"primaryZone,omitempty"`
	ObproxyClusterIDs    []int64             `json:"obproxyClusterIds,omitempty"`
	ObproxyUserName      string              `json:"obproxyUserName,omitempty"`
	ObproxyUserPassword  string              `json:"obproxyUserPassword,omitempty"`
	ArbitrationServiceID *int64              `json:"arbitrationServiceId,omitempty"`
	PrimaryClusterInfo   *PrimaryClusterInfo `json:"primaryClusterInfo,omitempty"`
	OversellingFactor    *int                `json:"oversellingFactor,omitempty"`
	CgroupEnabled        *bool               `json:"cgroupEnabled,omitempty"`
	CreateExtraTenant    bool                `json:"createExtraTenant,omitempty"`
	SupportObsBackup     bool                `json:"supportObsBackup,omitempty"`
	LoadType             string              `json:"loadType,omitempty"`
	CheckID              string              `json:"checkId,omitempty"`
	ClientToken          string              `json:"clientToken,omitempty"`
}

type ZoneParam struct {
	Name                   string  `json:"name"`
	IdcName                string  `json:"idcName"`
	Servers                []int64 `json:"servers"`
	RpmName                string  `json:"rpmName"`
	Architecture           string  `json:"architecture,omitempty"`
	PackageOperatingSystem string  `json:"packageOperatingSystem,omitempty"`
	RootServer             *int64  `json:"rootServer,omitempty"`
}

// ClusterAttributes uses PascalCase JSON keys to match OCP @JsonProperty annotations.
type ClusterAttributes struct {
	InstallPath         string `json:"InstallPath,omitempty"`
	RunPath             string `json:"RunPath,omitempty"`
	DataDiskPath        string `json:"DataDiskPath,omitempty"`
	LogDiskPath         string `json:"LogDiskPath,omitempty"`
	DiskPathStyle       string `json:"DiskPathStyle,omitempty"`
	OperatingSystemUser string `json:"OperatingSystemUser,omitempty"`
	SqlPort             *int   `json:"SqlPort,omitempty"`
	SvrPort             *int   `json:"SvrPort,omitempty"`
}

type PrimaryClusterInfo struct {
	ObClusterID     int64  `json:"obClusterId,omitempty"`
	RootSysPassword string `json:"rootSysPassword,omitempty"`
}

type KVParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cluster struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	FullVersion string `json:"fullVersion"`
}

type createClusterResp struct {
	ID        int64 `json:"id"`
	ClusterID int64 `json:"clusterId"`
}

// CreateCluster creates an OB cluster and returns (taskID, clusterID, error)
func (c *Client) CreateCluster(ctx context.Context, param CreateClusterParam) (int64, int64, error) {
	var resp createClusterResp
	if err := c.doRequest(ctx, "POST", "/api/v2/ob/clusters", param, &resp); err != nil {
		return 0, 0, err
	}
	if err := c.WaitForTask(ctx, resp.ID); err != nil {
		return resp.ID, resp.ClusterID, err
	}
	return resp.ID, resp.ClusterID, nil
}

func (c *Client) GetCluster(ctx context.Context, clusterID int64) (*Cluster, error) {
	var cl Cluster
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v2/ob/clusters/%d", clusterID), nil, &cl); err != nil {
		return nil, err
	}
	return &cl, nil
}

func (c *Client) DeleteCluster(ctx context.Context, clusterID int64) error {
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v2/ob/clusters/%d", clusterID), nil, &resp); err != nil {
		return err
	}
	return c.WaitForTask(ctx, resp.ID)
}

func (c *Client) ListClusters(ctx context.Context) ([]Cluster, error) {
	var result struct {
		Contents []Cluster `json:"contents"`
	}
	if err := c.doRequest(ctx, "GET", "/api/v2/ob/clusters", nil, &result); err != nil {
		return nil, err
	}
	return result.Contents, nil
}
