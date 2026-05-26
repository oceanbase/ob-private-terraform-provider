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

type arbitrationResource struct{ client *ocpclient.Client }

func NewArbitrationResource() resource.Resource { return &arbitrationResource{} }

type arbitrationModel struct {
	ID                   types.String `tfsdk:"id"`
	RpmName              types.String `tfsdk:"rpm_name"`
	HostID               types.Int64  `tfsdk:"host_id"`
	InstallPath          types.String `tfsdk:"install_path"`
	RunPath              types.String `tfsdk:"run_path"`
	SvrPort              types.Int64  `tfsdk:"svr_port"`
	RunUser              types.String `tfsdk:"run_user"`
	Description          types.String `tfsdk:"description"`
	ClogSymbolicLinkPath types.String `tfsdk:"clog_symbolic_link_path"`
	StartupParameters    []kvPair     `tfsdk:"startup_parameters"`
	ClientToken          types.String `tfsdk:"client_token"`
	Status               types.String `tfsdk:"status"`
}

func (r *arbitrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_arbitration_service"
}

func (r *arbitrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"rpm_name":                schema.StringAttribute{Required: true, Description: "RPM package full name"},
			"host_id":                 schema.Int64Attribute{Optional: true, Description: "Target host ID (>0)"},
			"install_path":            schema.StringAttribute{Optional: true, Description: "Install path, default /home/admin/oceanbase"},
			"run_path":                schema.StringAttribute{Optional: true, Description: "Run path, defaults to install_path"},
			"svr_port":                schema.Int64Attribute{Optional: true, Description: "Server port, default 2882"},
			"run_user":                schema.StringAttribute{Optional: true, Description: "Run user, default admin"},
			"description":             schema.StringAttribute{Optional: true},
			"clog_symbolic_link_path": schema.StringAttribute{Optional: true, Description: "Clog symbolic link path"},
			"client_token":            schema.StringAttribute{Optional: true, Description: "Idempotency token, length 8-64"},
			"status":                  schema.StringAttribute{Computed: true},
			"startup_parameters": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Startup parameter KV list",
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

func (r *arbitrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *arbitrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan arbitrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ok, err := r.client.IsArbitrationSupported(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to detect arbitration service support", err.Error())
		return
	}
	if !ok {
		resp.Diagnostics.AddError("Arbitration service not supported", "arbitration_service is only available in OceanBase Enterprise Edition")
		return
	}
	tflog.Info(ctx, "creating resource", map[string]interface{}{"type": "oceanbase_arbitration_service"})
	startupParams := make([]ocpclient.KVParam, 0, len(plan.StartupParameters))
	for _, p := range plan.StartupParameters {
		startupParams = append(startupParams, ocpclient.KVParam{
			Name:  p.Name.ValueString(),
			Value: p.Value.ValueString(),
		})
	}
	_, id, err := r.client.CreateArbitration(ctx, ocpclient.CreateArbitrationParam{
		HostID:               plan.HostID.ValueInt64(),
		RpmName:              plan.RpmName.ValueString(),
		InstallPath:          plan.InstallPath.ValueString(),
		RunPath:              plan.RunPath.ValueString(),
		SvrPort:              int(plan.SvrPort.ValueInt64()),
		RunUser:              plan.RunUser.ValueString(),
		Description:          plan.Description.ValueString(),
		ClogSymbolicLinkPath: plan.ClogSymbolicLinkPath.ValueString(),
		StartupParameters:    startupParams,
		ClientToken:          plan.ClientToken.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create arbitration service", err.Error())
		return
	}
	plan.ID = types.StringValue(strconv.FormatInt(id, 10))
	plan.Status = types.StringValue("CREATED")
	tflog.Info(ctx, "created resource", map[string]interface{}{"type": "oceanbase_arbitration_service", "id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *arbitrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state arbitrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "reading resource", map[string]interface{}{"type": "oceanbase_arbitration_service", "id": state.ID.ValueString()})
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid arbitration service ID", err.Error())
		return
	}
	a, err := r.client.GetArbitration(ctx, id)
	if err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read arbitration service", err.Error())
		return
	}
	state.Status = types.StringValue(a.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *arbitrationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Please recreate the resource to apply changes")
}

func (r *arbitrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state arbitrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "deleting resource", map[string]interface{}{"type": "oceanbase_arbitration_service", "id": state.ID.ValueString()})
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid arbitration service ID", err.Error())
		return
	}
	if err := r.client.DeleteArbitration(ctx, id); err != nil {
		var nfe *ocpclient.NotFoundError
		if errorAsNotFound(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete arbitration service", err.Error())
	}
}
