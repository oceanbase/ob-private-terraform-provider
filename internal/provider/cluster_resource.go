package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oceanbase/terraform-provider-oceanbase/internal/ocpclient"
)

type clusterResource struct {
	client *ocpclient.Client
}

func NewClusterResource() resource.Resource { return &clusterResource{} }

type clusterModel struct {
	ID                   types.String              `tfsdk:"id"`
	Name                 types.String              `tfsdk:"name"`
	Type                 types.String              `tfsdk:"type"`
	Password             types.String              `tfsdk:"password"`
	FullVersion          types.String              `tfsdk:"full_version"`
	ObClusterID          types.Int64               `tfsdk:"ob_cluster_id"`
	Zones                []clusterZoneModel        `tfsdk:"zones"`
	Attributes           *clusterAttributesModel   `tfsdk:"attributes"`
	StartupParameters    []kvPair                  `tfsdk:"startup_parameters"`
	PrimaryZone          types.String              `tfsdk:"primary_zone"`
	ObproxyClusterIDs    []int64                   `tfsdk:"obproxy_cluster_ids"`
	ObproxyUserName      types.String              `tfsdk:"obproxy_user_name"`
	ObproxyUserPassword  types.String              `tfsdk:"obproxy_user_password"`
	ArbitrationServiceID types.Int64               `tfsdk:"arbitration_service_id"`
	PrimaryClusterInfo   *primaryClusterInfoModel  `tfsdk:"primary_cluster_info"`
	OversellingFactor    types.Int64               `tfsdk:"overselling_factor"`
	CgroupEnabled        types.Bool                `tfsdk:"cgroup_enabled"`
	CreateExtraTenant    types.Bool                `tfsdk:"create_extra_tenant"`
	SupportObsBackup     types.Bool                `tfsdk:"support_obs_backup"`
	LoadType             types.String              `tfsdk:"load_type"`
	CheckID              types.String              `tfsdk:"check_id"`
	ClientToken          types.String              `tfsdk:"client_token"`
	Status               types.String              `tfsdk:"status"`
}

type clusterZoneModel struct {
	Name                   types.String `tfsdk:"name"`
	IdcName                types.String `tfsdk:"idc_name"`
	Servers                []int64      `tfsdk:"servers"`
	RpmName                types.String `tfsdk:"rpm_name"`
	Architecture           types.String `tfsdk:"architecture"`
	PackageOperatingSystem types.String `tfsdk:"package_operating_system"`
	RootServer             types.Int64  `tfsdk:"root_server"`
}

type clusterAttributesModel struct {
	InstallPath         types.String `tfsdk:"install_path"`
	RunPath             types.String `tfsdk:"run_path"`
	DataDiskPath        types.String `tfsdk:"data_disk_path"`
	LogDiskPath         types.String `tfsdk:"log_disk_path"`
	DiskPathStyle       types.String `tfsdk:"disk_path_style"`
	OperatingSystemUser types.String `tfsdk:"operating_system_user"`
	SqlPort             types.Int64  `tfsdk:"sql_port"`
	SvrPort             types.Int64  `tfsdk:"svr_port"`
}

type primaryClusterInfoModel struct {
	ObClusterID     types.Int64  `tfsdk:"ob_cluster_id"`
	RootSysPassword types.String `tfsdk:"root_sys_password"`
}

type kvPair struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ob_cluster"
}

func (r *clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":                   schema.StringAttribute{Required: true, Description: "Cluster name, must match ^[a-zA-Z][a-zA-Z_0-9]{1,47}$"},
			"type":                   schema.StringAttribute{Required: true, Description: "PRIMARY or STANDBY"},
			"password":               schema.StringAttribute{Optional: true, Sensitive: true, Description: "root@sys password"},
			"full_version":           schema.StringAttribute{Optional: true, Description: "Full OB version, e.g. 4.2.1.0-100000242023112107"},
			"ob_cluster_id":          schema.Int64Attribute{Optional: true, Description: "Upstream OB cluster ID"},
			"primary_zone":           schema.StringAttribute{Optional: true, Description: "Primary zone, e.g. zone1;zone2,zone3"},
			"obproxy_cluster_ids":    schema.ListAttribute{Optional: true, ElementType: types.Int64Type, Description: "Associated OBProxy cluster IDs"},
			"obproxy_user_name":      schema.StringAttribute{Optional: true, Description: "OBProxy user name, default proxyro"},
			"obproxy_user_password":  schema.StringAttribute{Optional: true, Sensitive: true, Description: "OBProxy user password"},
			"arbitration_service_id": schema.Int64Attribute{Optional: true, Description: "Arbitration service ID"},
			"overselling_factor":     schema.Int64Attribute{Optional: true, Description: "Resource overselling factor in [100, 200]"},
			"cgroup_enabled":         schema.BoolAttribute{Optional: true, Description: "Whether to enable cgroup, default true"},
			"create_extra_tenant":    schema.BoolAttribute{Optional: true, Description: "Whether to create extra tenants, default false"},
			"support_obs_backup":     schema.BoolAttribute{Optional: true, Description: "Whether to support Huawei Cloud OBS backup/restore"},
			"load_type":              schema.StringAttribute{Optional: true, Description: "Cluster load type"},
			"check_id":               schema.StringAttribute{Optional: true, Description: "Pre-check ID"},
			"client_token":           schema.StringAttribute{Optional: true, Description: "Idempotency token, length 8-64"},
			"status":                 schema.StringAttribute{Computed: true},
			"zones": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     schema.StringAttribute{Required: true},
						"idc_name": schema.StringAttribute{Required: true},
						"servers": schema.ListAttribute{
							Required:    true,
							ElementType: types.Int64Type,
							Description: "Host ID list",
						},
						"rpm_name":                 schema.StringAttribute{Required: true, Description: "RPM package full file name"},
						"architecture":             schema.StringAttribute{Optional: true, Description: "x86_64 or aarch64"},
						"package_operating_system": schema.StringAttribute{Optional: true, Description: "el7 or el8"},
						"root_server":              schema.Int64Attribute{Optional: true, Description: "Root server host ID"},
					},
				},
			},
			"attributes": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Cluster path and port attributes",
				Attributes: map[string]schema.Attribute{
					"install_path":          schema.StringAttribute{Optional: true, Description: "Install path, default /home/admin/oceanbase"},
					"run_path":              schema.StringAttribute{Optional: true, Description: "Run path"},
					"data_disk_path":        schema.StringAttribute{Optional: true, Description: "Data disk path, default /data/1"},
					"log_disk_path":         schema.StringAttribute{Optional: true, Description: "Log disk path, default /data/log1"},
					"disk_path_style":       schema.StringAttribute{Optional: true, Description: "Disk path style"},
					"operating_system_user": schema.StringAttribute{Optional: true, Description: "Host OS user"},
					"sql_port":              schema.Int64Attribute{Optional: true, Description: "SQL port, default 2881"},
					"svr_port":              schema.Int64Attribute{Optional: true, Description: "RPC port, default 2882"},
				},
			},
			"primary_cluster_info": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Primary cluster info, required for STANDBY type",
				Attributes: map[string]schema.Attribute{
					"ob_cluster_id":     schema.Int64Attribute{Optional: true},
					"root_sys_password": schema.StringAttribute{Optional: true, Sensitive: true},
				},
			},
			"startup_parameters": schema.ListNestedAttribute{
				Optional:    true,
				Description: "OB startup parameter KV list",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":  schema.StringAttribute{Required: true},
						"value": schema.StringAttribute{Required: true},
					},
				},
			},
		},
	}
}

func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ocpclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *ocpclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zones := make([]ocpclient.ZoneParam, 0, len(plan.Zones))
	for _, z := range plan.Zones {
		zp := ocpclient.ZoneParam{
			Name:                   z.Name.ValueString(),
			IdcName:                z.IdcName.ValueString(),
			Servers:                z.Servers,
			RpmName:                z.RpmName.ValueString(),
			Architecture:           z.Architecture.ValueString(),
			PackageOperatingSystem: z.PackageOperatingSystem.ValueString(),
		}
		if !z.RootServer.IsNull() {
			v := z.RootServer.ValueInt64()
			zp.RootServer = &v
		}
		zones = append(zones, zp)
	}

	param := ocpclient.CreateClusterParam{
		Name:                plan.Name.ValueString(),
		Type:                plan.Type.ValueString(),
		Password:            plan.Password.ValueString(),
		FullVersion:         plan.FullVersion.ValueString(),
		Zones:               zones,
		PrimaryZone:         plan.PrimaryZone.ValueString(),
		ObproxyClusterIDs:   plan.ObproxyClusterIDs,
		ObproxyUserName:     plan.ObproxyUserName.ValueString(),
		ObproxyUserPassword: plan.ObproxyUserPassword.ValueString(),
		CreateExtraTenant:   plan.CreateExtraTenant.ValueBool(),
		SupportObsBackup:    plan.SupportObsBackup.ValueBool(),
		LoadType:            plan.LoadType.ValueString(),
		CheckID:             plan.CheckID.ValueString(),
		ClientToken:         plan.ClientToken.ValueString(),
	}
	if !plan.ObClusterID.IsNull() {
		v := plan.ObClusterID.ValueInt64()
		param.ObClusterID = &v
	}
	if !plan.ArbitrationServiceID.IsNull() {
		v := plan.ArbitrationServiceID.ValueInt64()
		param.ArbitrationServiceID = &v
	}
	if !plan.OversellingFactor.IsNull() {
		v := int(plan.OversellingFactor.ValueInt64())
		param.OversellingFactor = &v
	}
	if !plan.CgroupEnabled.IsNull() {
		v := plan.CgroupEnabled.ValueBool()
		param.CgroupEnabled = &v
	}
	if plan.Attributes != nil {
		attrs := &ocpclient.ClusterAttributes{
			InstallPath:         plan.Attributes.InstallPath.ValueString(),
			RunPath:             plan.Attributes.RunPath.ValueString(),
			DataDiskPath:        plan.Attributes.DataDiskPath.ValueString(),
			LogDiskPath:         plan.Attributes.LogDiskPath.ValueString(),
			DiskPathStyle:       plan.Attributes.DiskPathStyle.ValueString(),
			OperatingSystemUser: plan.Attributes.OperatingSystemUser.ValueString(),
		}
		if !plan.Attributes.SqlPort.IsNull() {
			v := int(plan.Attributes.SqlPort.ValueInt64())
			attrs.SqlPort = &v
		}
		if !plan.Attributes.SvrPort.IsNull() {
			v := int(plan.Attributes.SvrPort.ValueInt64())
			attrs.SvrPort = &v
		}
		param.Attributes = attrs
	}
	if plan.PrimaryClusterInfo != nil {
		param.PrimaryClusterInfo = &ocpclient.PrimaryClusterInfo{
			ObClusterID:     plan.PrimaryClusterInfo.ObClusterID.ValueInt64(),
			RootSysPassword: plan.PrimaryClusterInfo.RootSysPassword.ValueString(),
		}
	}
	for _, p := range plan.StartupParameters {
		param.StartupParameters = append(param.StartupParameters, ocpclient.KVParam{
			Name:  p.Name.ValueString(),
			Value: p.Value.ValueString(),
		})
	}

	tflog.Info(ctx, "creating resource", map[string]interface{}{"type": "oceanbase_ob_cluster", "name": plan.Name.ValueString()})
	_, clusterID, err := r.client.CreateCluster(ctx, param)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cluster", err.Error())
		return
	}
	plan.ID = types.StringValue(strconv.FormatInt(clusterID, 10))
	plan.Status = types.StringValue("CREATED")
	tflog.Info(ctx, "created resource", map[string]interface{}{"type": "oceanbase_ob_cluster", "id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "reading resource", map[string]interface{}{"type": "oceanbase_ob_cluster", "id": state.ID.ValueString()})
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster ID", err.Error())
		return
	}
	cl, err := r.client.GetCluster(ctx, id)
	if err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read cluster", err.Error())
		return
	}
	state.Status = types.StringValue(cl.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clusterResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "oceanbase_ob_cluster does not support in-place updates in MVP, please recreate the resource")
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "deleting resource", map[string]interface{}{"type": "oceanbase_ob_cluster", "id": state.ID.ValueString()})
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster ID", err.Error())
		return
	}
	if err := r.client.DeleteCluster(ctx, id); err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete cluster", err.Error())
	}
}
