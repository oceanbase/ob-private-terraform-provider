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
	"github.com/oceanbase/terraform-provider-oceanbase/internal/ocpclient"
)

type obproxyResource struct{ client *ocpclient.Client }

func NewObproxyResource() resource.Resource { return &obproxyResource{} }

type obproxyModel struct {
	ID                  types.String         `tfsdk:"id"`
	Name                types.String         `tfsdk:"name"`
	Password            types.String         `tfsdk:"password"`
	ProxyroPassword     types.String         `tfsdk:"proxyro_password"`
	WorkMode            types.String         `tfsdk:"work_mode"`
	InstallPath         types.String         `tfsdk:"install_path"`
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
}

type obLinkModel struct {
	ClusterName types.String `tfsdk:"cluster_name"`
	ObClusterID types.Int64  `tfsdk:"ob_cluster_id"`
}

func (r *obproxyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_obproxy_cluster"
}

func (r *obproxyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":             schema.StringAttribute{Required: true},
			"password":         schema.StringAttribute{Optional: true, Sensitive: true},
			"proxyro_password": schema.StringAttribute{Optional: true, Sensitive: true},
			"work_mode":        schema.StringAttribute{Optional: true},
			"install_path":     schema.StringAttribute{Optional: true},
			"client_token":     schema.StringAttribute{Optional: true},
			"status":           schema.StringAttribute{Computed: true},
			"obproxy_install_param": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"host_ids":      schema.ListAttribute{Optional: true, ElementType: types.Int64Type},
					"version":       schema.StringAttribute{Optional: true},
					"sql_port":      schema.Int64Attribute{Optional: true},
					"exporter_port": schema.Int64Attribute{Optional: true},
				},
			},
			"ob_links": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_name":  schema.StringAttribute{Required: true},
						"ob_cluster_id": schema.Int64Attribute{Optional: true},
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
		resp.Diagnostics.AddError("Provider 数据类型异常", fmt.Sprintf("实际收到 %T", req.ProviderData))
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
		WorkMode:        plan.WorkMode.ValueString(),
		InstallPath:     plan.InstallPath.ValueString(),
		ClientToken:     plan.ClientToken.ValueString(),
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
		param.ObproxyInstallParam = ip
	}
	for _, l := range plan.ObLinks {
		param.ObLinks = append(param.ObLinks, ocpclient.ObLinkParam{
			ClusterName: l.ClusterName.ValueString(),
			ObClusterID: l.ObClusterID.ValueInt64(),
		})
	}
	_, id, err := r.client.CreateObproxy(ctx, param)
	if err != nil {
		resp.Diagnostics.AddError("创建 OBProxy 失败", err.Error())
		return
	}
	plan.ID = types.StringValue(strconv.FormatInt(id, 10))
	plan.Status = types.StringValue("CREATED")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *obproxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state obproxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("OBProxy ID 无效", err.Error())
		return
	}
	o, err := r.client.GetObproxy(ctx, id)
	if err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("读取 OBProxy 失败", err.Error())
		return
	}
	state.Status = types.StringValue(o.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *obproxyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("不支持更新操作", "请重建资源以应用变更")
}

func (r *obproxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state obproxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("OBProxy ID 无效", err.Error())
		return
	}
	if err := r.client.DeleteObproxy(ctx, id); err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("删除 OBProxy 失败", err.Error())
	}
}
