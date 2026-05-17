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

type arbitrationResource struct{ client *ocpclient.Client }

func NewArbitrationResource() resource.Resource { return &arbitrationResource{} }

type arbitrationModel struct {
	ID          types.String `tfsdk:"id"`
	RpmName     types.String `tfsdk:"rpm_name"`
	HostID      types.Int64  `tfsdk:"host_id"`
	InstallPath types.String `tfsdk:"install_path"`
	SvrPort     types.Int64  `tfsdk:"svr_port"`
	RunUser     types.String `tfsdk:"run_user"`
	Description types.String `tfsdk:"description"`
	ClientToken types.String `tfsdk:"client_token"`
	Status      types.String `tfsdk:"status"`
}

func (r *arbitrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_arbitration_service"
}

func (r *arbitrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"rpm_name":     schema.StringAttribute{Required: true},
			"host_id":      schema.Int64Attribute{Optional: true},
			"install_path": schema.StringAttribute{Optional: true},
			"svr_port":     schema.Int64Attribute{Optional: true},
			"run_user":     schema.StringAttribute{Optional: true},
			"description":  schema.StringAttribute{Optional: true},
			"client_token": schema.StringAttribute{Optional: true},
			"status":       schema.StringAttribute{Computed: true},
		},
	}
}

func (r *arbitrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *arbitrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan arbitrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ok, err := r.client.IsArbitrationSupported(ctx)
	if err != nil {
		resp.Diagnostics.AddError("探测仲裁服务支持性失败", err.Error())
		return
	}
	if !ok {
		resp.Diagnostics.AddError("当前 OCP 不支持仲裁服务", "arbitration_service 仅在 OceanBase 企业版中可用")
		return
	}
	_, id, err := r.client.CreateArbitration(ctx, ocpclient.CreateArbitrationParam{
		HostID:      plan.HostID.ValueInt64(),
		RpmName:     plan.RpmName.ValueString(),
		InstallPath: plan.InstallPath.ValueString(),
		SvrPort:     int(plan.SvrPort.ValueInt64()),
		RunUser:     plan.RunUser.ValueString(),
		Description: plan.Description.ValueString(),
		ClientToken: plan.ClientToken.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("创建仲裁服务失败", err.Error())
		return
	}
	plan.ID = types.StringValue(strconv.FormatInt(id, 10))
	plan.Status = types.StringValue("CREATED")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *arbitrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state arbitrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("仲裁服务 ID 无效", err.Error())
		return
	}
	a, err := r.client.GetArbitration(ctx, id)
	if err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("读取仲裁服务失败", err.Error())
		return
	}
	state.Status = types.StringValue(a.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *arbitrationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("不支持更新操作", "请重建资源以应用变更")
}

func (r *arbitrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state arbitrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("仲裁服务 ID 无效", err.Error())
		return
	}
	if err := r.client.DeleteArbitration(ctx, id); err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("删除仲裁服务失败", err.Error())
	}
}
