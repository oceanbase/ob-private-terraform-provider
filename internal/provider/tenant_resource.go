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

type tenantResource struct{ client *ocpclient.Client }

func NewTenantResource() resource.Resource { return &tenantResource{} }

type tenantModel struct {
	ID                types.String      `tfsdk:"id"`
	ClusterID         types.Int64       `tfsdk:"cluster_id"`
	Name              types.String      `tfsdk:"name"`
	RootPassword      types.String      `tfsdk:"root_password"`
	Mode              types.String      `tfsdk:"mode"`
	PrimaryZone       types.String      `tfsdk:"primary_zone"`
	Charset           types.String      `tfsdk:"charset"`
	Collation         types.String      `tfsdk:"collation"`
	Whitelist         types.String      `tfsdk:"whitelist"`
	Description       types.String      `tfsdk:"description"`
	EnableArbitration types.Bool        `tfsdk:"enable_arbitration"`
	ClientToken       types.String      `tfsdk:"client_token"`
	Zones             []tenantZoneModel `tfsdk:"zones"`
	Parameters        []kvModel         `tfsdk:"parameters"`
	Status            types.String      `tfsdk:"status"`
}

type tenantZoneModel struct {
	Name         types.String      `tfsdk:"name"`
	ReplicaType  types.String      `tfsdk:"replica_type"`
	ResourcePool resourcePoolModel `tfsdk:"resource_pool"`
}

type resourcePoolModel struct {
	UnitSpecName types.String `tfsdk:"unit_spec_name"`
	UnitCount    types.Int64  `tfsdk:"unit_count"`
}

type kvModel struct {
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
	ParameterType types.String `tfsdk:"parameter_type"`
}

func (r *tenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ob_tenant"
}

func (r *tenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"cluster_id":         schema.Int64Attribute{Required: true},
			"name":               schema.StringAttribute{Required: true},
			"root_password":      schema.StringAttribute{Required: true, Sensitive: true},
			"mode":               schema.StringAttribute{Optional: true},
			"primary_zone":       schema.StringAttribute{Optional: true},
			"charset":            schema.StringAttribute{Optional: true},
			"collation":          schema.StringAttribute{Optional: true},
			"whitelist":          schema.StringAttribute{Optional: true},
			"description":        schema.StringAttribute{Optional: true},
			"enable_arbitration": schema.BoolAttribute{Optional: true},
			"client_token":       schema.StringAttribute{Optional: true},
			"status":             schema.StringAttribute{Computed: true},
			"zones": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":         schema.StringAttribute{Required: true},
						"replica_type": schema.StringAttribute{Required: true},
						"resource_pool": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"unit_spec_name": schema.StringAttribute{Required: true},
								"unit_count":     schema.Int64Attribute{Required: true},
							},
						},
					},
				},
			},
			"parameters": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":           schema.StringAttribute{Required: true},
						"value":          schema.StringAttribute{Required: true},
						"parameter_type": schema.StringAttribute{Optional: true},
					},
				},
			},
		},
	}
}

func (r *tenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zones := make([]ocpclient.TenantZoneParam, 0, len(plan.Zones))
	for _, z := range plan.Zones {
		zones = append(zones, ocpclient.TenantZoneParam{
			Name:        z.Name.ValueString(),
			ReplicaType: z.ReplicaType.ValueString(),
			ResourcePool: ocpclient.PoolParam{
				UnitSpecName: z.ResourcePool.UnitSpecName.ValueString(),
				UnitCount:    z.ResourcePool.UnitCount.ValueInt64(),
			},
		})
	}
	params := make([]ocpclient.TenantParameterParam, 0, len(plan.Parameters))
	for _, p := range plan.Parameters {
		params = append(params, ocpclient.TenantParameterParam{
			Name: p.Name.ValueString(), Value: p.Value.ValueString(), ParameterType: p.ParameterType.ValueString(),
		})
	}
	_, tenantID, err := r.client.CreateTenant(ctx, plan.ClusterID.ValueInt64(), ocpclient.CreateTenantParam{
		Name:              plan.Name.ValueString(),
		Mode:              plan.Mode.ValueString(),
		PrimaryZone:       plan.PrimaryZone.ValueString(),
		Charset:           plan.Charset.ValueString(),
		Collation:         plan.Collation.ValueString(),
		Whitelist:         plan.Whitelist.ValueString(),
		Description:       plan.Description.ValueString(),
		RootPassword:      plan.RootPassword.ValueString(),
		EnableArbitration: plan.EnableArbitration.ValueBool(),
		Zones:             zones,
		Parameters:        params,
		ClientToken:       plan.ClientToken.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("创建租户失败", err.Error())
		return
	}
	plan.ID = types.StringValue(strconv.FormatInt(tenantID, 10))
	plan.Status = types.StringValue("CREATED")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("租户 ID 无效", err.Error())
		return
	}
	t, err := r.client.GetTenant(ctx, state.ClusterID.ValueInt64(), id)
	if err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("读取租户失败", err.Error())
		return
	}
	state.Status = types.StringValue(t.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tenantResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("不支持更新操作", "请重建资源以应用变更")
}

func (r *tenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("租户 ID 无效", err.Error())
		return
	}
	if err := r.client.DeleteTenant(ctx, state.ClusterID.ValueInt64(), id); err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("删除租户失败", err.Error())
	}
}
