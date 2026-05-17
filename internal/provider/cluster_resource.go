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

type clusterResource struct {
	client *ocpclient.Client
}

func NewClusterResource() resource.Resource { return &clusterResource{} }

type clusterModel struct {
	ID          types.String       `tfsdk:"id"`
	Name        types.String       `tfsdk:"name"`
	Type        types.String       `tfsdk:"type"`
	Password    types.String       `tfsdk:"password"`
	FullVersion types.String       `tfsdk:"full_version"`
	Zones       []clusterZoneModel `tfsdk:"zones"`
	PrimaryZone types.String       `tfsdk:"primary_zone"`
	ClientToken types.String       `tfsdk:"client_token"`
	Status      types.String       `tfsdk:"status"`
}

type clusterZoneModel struct {
	Name                   types.String `tfsdk:"name"`
	IdcName                types.String `tfsdk:"idc_name"`
	Servers                []int64      `tfsdk:"servers"`
	RpmName                types.String `tfsdk:"rpm_name"`
	Architecture           types.String `tfsdk:"architecture"`
	PackageOperatingSystem types.String `tfsdk:"package_operating_system"`
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
			"name":         schema.StringAttribute{Required: true},
			"type":         schema.StringAttribute{Required: true, Description: "PRIMARY 或 STANDBY"},
			"password":     schema.StringAttribute{Optional: true, Sensitive: true},
			"full_version": schema.StringAttribute{Optional: true},
			"primary_zone": schema.StringAttribute{Optional: true},
			"client_token": schema.StringAttribute{Optional: true},
			"status":       schema.StringAttribute{Computed: true},
			"zones": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     schema.StringAttribute{Required: true},
						"idc_name": schema.StringAttribute{Required: true},
						"servers": schema.ListAttribute{
							Required:    true,
							ElementType: types.Int64Type,
						},
						"rpm_name":                 schema.StringAttribute{Required: true},
						"architecture":             schema.StringAttribute{Optional: true},
						"package_operating_system": schema.StringAttribute{Optional: true},
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
		resp.Diagnostics.AddError("Provider 数据类型异常", fmt.Sprintf("实际收到 %T", req.ProviderData))
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
		zones = append(zones, ocpclient.ZoneParam{
			Name:                   z.Name.ValueString(),
			IdcName:                z.IdcName.ValueString(),
			Servers:                z.Servers,
			RpmName:                z.RpmName.ValueString(),
			Architecture:           z.Architecture.ValueString(),
			PackageOperatingSystem: z.PackageOperatingSystem.ValueString(),
		})
	}

	param := ocpclient.CreateClusterParam{
		Name:        plan.Name.ValueString(),
		Type:        plan.Type.ValueString(),
		Password:    plan.Password.ValueString(),
		FullVersion: plan.FullVersion.ValueString(),
		Zones:       zones,
		PrimaryZone: plan.PrimaryZone.ValueString(),
		ClientToken: plan.ClientToken.ValueString(),
	}
	_, clusterID, err := r.client.CreateCluster(ctx, param)
	if err != nil {
		resp.Diagnostics.AddError("创建集群失败", err.Error())
		return
	}
	plan.ID = types.StringValue(strconv.FormatInt(clusterID, 10))
	plan.Status = types.StringValue("CREATED")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("集群 ID 无效", err.Error())
		return
	}
	cl, err := r.client.GetCluster(ctx, id)
	if err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("读取集群失败", err.Error())
		return
	}
	state.Status = types.StringValue(cl.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clusterResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("不支持更新操作", "oceanbase_ob_cluster 在 MVP 阶段不支持原地更新，请重建资源")
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("集群 ID 无效", err.Error())
		return
	}
	if err := r.client.DeleteCluster(ctx, id); err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("删除集群失败", err.Error())
	}
}
