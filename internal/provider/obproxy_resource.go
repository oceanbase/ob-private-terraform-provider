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

type obproxyResource struct{ client *ocpclient.Client }

func NewObproxyResource() resource.Resource { return &obproxyResource{} }

type obproxyModel struct {
	ID                  types.String         `tfsdk:"id"`
	Name                types.String         `tfsdk:"name"`
	Password            types.String         `tfsdk:"password"`
	ProxyroPassword     types.String         `tfsdk:"proxyro_password"`
	Address             types.String         `tfsdk:"address"`
	Port                types.Int64          `tfsdk:"port"`
	WorkMode            types.String         `tfsdk:"work_mode"`
	InstallPath         types.String         `tfsdk:"install_path"`
	RunPath             types.String         `tfsdk:"run_path"`
	RunUser             types.String         `tfsdk:"run_user"`
	StartupParameters   []kvPair             `tfsdk:"startup_parameters"`
	Parameters          []kvPair             `tfsdk:"parameters"`
	ClientToken         types.String         `tfsdk:"client_token"`
	Status              types.String         `tfsdk:"status"`
	ObproxyInstallParam *obproxyInstallModel `tfsdk:"obproxy_install_param"`
	ObLinks             []obLinkModel        `tfsdk:"ob_links"`
}

type obproxyInstallModel struct {
	HostIDs      []int64      `tfsdk:"host_ids"`
	Version      types.String `tfsdk:"version"`
	SqlPort      types.Int64  `tfsdk:"sql_port"`
	ExporterPort types.Int64  `tfsdk:"exporter_port"`
	RpcPort      types.Int64  `tfsdk:"rpc_port"`
}

type obLinkModel struct {
	ClusterName types.String `tfsdk:"cluster_name"`
	ObClusterID types.Int64  `tfsdk:"ob_cluster_id"`
	Username    types.String `tfsdk:"username"`
}

func (r *obproxyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_obproxy_cluster"
}

func (r *obproxyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":             schema.StringAttribute{Required: true},
			"password":         schema.StringAttribute{Optional: true, Sensitive: true, Description: "root@proxysys password"},
			"proxyro_password": schema.StringAttribute{Optional: true, Sensitive: true, Description: "proxyro user password"},
			"address":          schema.StringAttribute{Optional: true, Description: "Cluster access address"},
			"port":             schema.Int64Attribute{Optional: true, Description: "Cluster access port"},
			"work_mode":        schema.StringAttribute{Optional: true, Description: "CONFIG_URL (default), RS_LIST, or OB_SHARDING"},
			"install_path":     schema.StringAttribute{Optional: true, Description: "Install path"},
			"run_path":         schema.StringAttribute{Optional: true, Description: "Run path"},
			"run_user":         schema.StringAttribute{Optional: true, Description: "Run user"},
			"client_token":     schema.StringAttribute{Optional: true, Description: "Idempotency token, length 8-64"},
			"status":           schema.StringAttribute{Computed: true},
			"obproxy_install_param": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"host_ids":      schema.ListAttribute{Optional: true, ElementType: types.Int64Type},
					"version":       schema.StringAttribute{Optional: true, Description: "Full RPM file name"},
					"sql_port":      schema.Int64Attribute{Optional: true},
					"exporter_port": schema.Int64Attribute{Optional: true},
					"rpc_port":      schema.Int64Attribute{Optional: true, Description: "RPC port (>=4.3.0)"},
				},
			},
			"ob_links": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_name":  schema.StringAttribute{Required: true},
						"ob_cluster_id": schema.Int64Attribute{Optional: true},
						"username":      schema.StringAttribute{Optional: true},
					},
				},
			},
			"startup_parameters": schema.ListNestedAttribute{
				Optional:    true,
				Description: "OBProxy startup parameters",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":  schema.StringAttribute{Required: true},
						"value": schema.StringAttribute{Required: true},
					},
				},
			},
			"parameters": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Post-startup parameters applied after OBProxy starts",
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

func (r *obproxyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*ocpclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *ocpclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *obproxyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan obproxyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	param := ocpclient.CreateObproxyParam{
		Name:            plan.Name.ValueString(),
		Password:        plan.Password.ValueString(),
		ProxyroPassword: plan.ProxyroPassword.ValueString(),
		Address:         plan.Address.ValueString(),
		WorkMode:        plan.WorkMode.ValueString(),
		InstallPath:     plan.InstallPath.ValueString(),
		RunPath:         plan.RunPath.ValueString(),
		RunUser:         plan.RunUser.ValueString(),
		ClientToken:     plan.ClientToken.ValueString(),
	}
	if !plan.Port.IsNull() {
		v := int(plan.Port.ValueInt64())
		param.Port = &v
	}
	if plan.ObproxyInstallParam != nil {
		ip := &ocpclient.InstallObproxyParam{
			HostIDs: plan.ObproxyInstallParam.HostIDs,
			Version: plan.ObproxyInstallParam.Version.ValueString(),
		}
		if !plan.ObproxyInstallParam.SqlPort.IsNull() {
			v := int(plan.ObproxyInstallParam.SqlPort.ValueInt64())
			ip.SqlPort = &v
		}
		if !plan.ObproxyInstallParam.ExporterPort.IsNull() {
			v := int(plan.ObproxyInstallParam.ExporterPort.ValueInt64())
			ip.ExporterPort = &v
		}
		if !plan.ObproxyInstallParam.RpcPort.IsNull() {
			v := int(plan.ObproxyInstallParam.RpcPort.ValueInt64())
			ip.RpcPort = &v
		}
		param.ObproxyInstallParam = ip
	}
	for _, l := range plan.ObLinks {
		param.ObLinks = append(param.ObLinks, ocpclient.ObLinkParam{
			ClusterName: l.ClusterName.ValueString(),
			ObClusterID: l.ObClusterID.ValueInt64(),
			Username:    l.Username.ValueString(),
		})
	}
	for _, p := range plan.StartupParameters {
		param.StartupParameters = append(param.StartupParameters, ocpclient.KVParam{
			Name:  p.Name.ValueString(),
			Value: p.Value.ValueString(),
		})
	}
	for _, p := range plan.Parameters {
		param.Parameters = append(param.Parameters, ocpclient.KVParam{
			Name:  p.Name.ValueString(),
			Value: p.Value.ValueString(),
		})
	}
	tflog.Info(ctx, "creating resource", map[string]interface{}{"type": "oceanbase_obproxy_cluster", "name": plan.Name.ValueString()})
	_, id, err := r.client.CreateObproxy(ctx, param)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create OBProxy cluster", err.Error())
		return
	}
	plan.ID = types.StringValue(strconv.FormatInt(id, 10))
	plan.Status = types.StringValue("CREATED")
	tflog.Info(ctx, "created resource", map[string]interface{}{"type": "oceanbase_obproxy_cluster", "id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *obproxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state obproxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "reading resource", map[string]interface{}{"type": "oceanbase_obproxy_cluster", "id": state.ID.ValueString()})
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid OBProxy cluster ID", err.Error())
		return
	}
	o, err := r.client.GetObproxy(ctx, id)
	if err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read OBProxy cluster", err.Error())
		return
	}
	state.Status = types.StringValue(o.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *obproxyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Please recreate the resource to apply changes")
}

func (r *obproxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state obproxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "deleting resource", map[string]interface{}{"type": "oceanbase_obproxy_cluster", "id": state.ID.ValueString()})
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid OBProxy cluster ID", err.Error())
		return
	}
	if err := r.client.DeleteObproxy(ctx, id); err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete OBProxy cluster", err.Error())
	}
}
