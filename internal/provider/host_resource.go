package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oceanbase/terraform-provider-oceanbase/internal/ocpclient"
)

type hostResource struct{ client *ocpclient.Client }

func NewHostResource() resource.Resource { return &hostResource{} }

type hostModel struct {
	ID                  types.String     `tfsdk:"id"`
	Hosts               []hostBasicModel `tfsdk:"hosts"`
	SshPort             types.Int64      `tfsdk:"ssh_port"`
	Kind                types.String     `tfsdk:"kind"`
	PublishPorts        []string         `tfsdk:"publish_ports"`
	IdcID               types.Int64      `tfsdk:"idc_id"`
	TypeID              types.Int64      `tfsdk:"type_id"`
	CredentialID        types.Int64      `tfsdk:"credential_id"`
	Alias               types.String     `tfsdk:"alias"`
	Description         types.String     `tfsdk:"description"`
	MgragentPort        types.Int64      `tfsdk:"mgragent_port"`
	MonagentPort        types.Int64      `tfsdk:"monagent_port"`
	MonagentCpuQuota    types.Float64    `tfsdk:"monagent_cpu_quota"`
	MonagentMemoryQuota types.Int64      `tfsdk:"monagent_memory_quota"`
	HostIDs             []int64          `tfsdk:"host_ids"`
}

type hostBasicModel struct {
	InnerIpAddress types.String `tfsdk:"inner_ip_address"`
	SerialNumber   types.String `tfsdk:"serial_number"`
}

func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                    schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ssh_port":              schema.Int64Attribute{Required: true},
			"kind":                  schema.StringAttribute{Required: true, Description: "DEDICATED_PHYSICAL_MACHINE, DEDICATED_CONTAINER, or DEDICATED_ECS"},
			"publish_ports":         schema.ListAttribute{Optional: true, ElementType: types.StringType, Description: "Published ports, only for DEDICATED_CONTAINER"},
			"idc_id":                schema.Int64Attribute{Required: true},
			"type_id":               schema.Int64Attribute{Required: true},
			"credential_id":         schema.Int64Attribute{Required: true},
			"alias":                 schema.StringAttribute{Optional: true},
			"description":           schema.StringAttribute{Optional: true},
			"mgragent_port":         schema.Int64Attribute{Optional: true, Description: "Manager agent port"},
			"monagent_port":         schema.Int64Attribute{Optional: true, Description: "Monitor agent port"},
			"monagent_cpu_quota":    schema.Float64Attribute{Optional: true, Description: "Monitor agent CPU quota, default 1.0"},
			"monagent_memory_quota": schema.Int64Attribute{Optional: true, Description: "Monitor agent memory quota in GiB, default 2"},
			"host_ids":              schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
			"hosts": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"inner_ip_address": schema.StringAttribute{Required: true},
						"serial_number":    schema.StringAttribute{Optional: true},
					},
				},
			},
		},
	}
}

func (r *hostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	basics := make([]ocpclient.HostBasicData, 0, len(plan.Hosts))
	for _, h := range plan.Hosts {
		basics = append(basics, ocpclient.HostBasicData{
			InnerIpAddress: h.InnerIpAddress.ValueString(),
			SerialNumber:   h.SerialNumber.ValueString(),
		})
	}
	createParam := ocpclient.BatchCreateHostParam{
		HostBasicDataList: basics,
		SshPort:           int(plan.SshPort.ValueInt64()),
		Kind:              plan.Kind.ValueString(),
		PublishPorts:      plan.PublishPorts,
		IdcID:             plan.IdcID.ValueInt64(),
		TypeID:            plan.TypeID.ValueInt64(),
		CredentialID:      plan.CredentialID.ValueInt64(),
		Alias:             plan.Alias.ValueString(),
		Description:       plan.Description.ValueString(),
	}
	if !plan.MgragentPort.IsNull() {
		v := int(plan.MgragentPort.ValueInt64())
		createParam.MgragentPort = &v
	}
	if !plan.MonagentPort.IsNull() {
		v := int(plan.MonagentPort.ValueInt64())
		createParam.MonagentPort = &v
	}
	if !plan.MonagentCpuQuota.IsNull() {
		v := plan.MonagentCpuQuota.ValueFloat64()
		createParam.MonagentCpuQuota = &v
	}
	if !plan.MonagentMemoryQuota.IsNull() {
		v := int(plan.MonagentMemoryQuota.ValueInt64())
		createParam.MonagentMemoryQuota = &v
	}
	tflog.Info(ctx, "creating resource", map[string]interface{}{"type": "oceanbase_host", "host_count": len(basics)})
	_, ids, err := r.client.BatchCreateHost(ctx, createParam)
	if err != nil {
		resp.Diagnostics.AddError("Failed to batch create hosts", err.Error())
		return
	}
	plan.ID = types.StringValue(uuid.NewString())
	plan.HostIDs = ids
	tflog.Info(ctx, "created resource", map[string]interface{}{"type": "oceanbase_host", "id": plan.ID.ValueString(), "host_ids": ids})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "reading resource", map[string]interface{}{"type": "oceanbase_host", "id": state.ID.ValueString()})
	alive := make([]int64, 0, len(state.HostIDs))
	for _, id := range state.HostIDs {
		_, err := r.client.GetHost(ctx, id)
		if err != nil {
			var nfe *ocpclient.NotFoundError
			if errorAsNotFound(err, &nfe) {
				continue
			}
			resp.Diagnostics.AddError("Failed to read host", err.Error())
			return
		}
		alive = append(alive, id)
	}
	if len(alive) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}
	state.HostIDs = alive
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hostResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Please recreate the resource to apply changes")
}

func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "deleting resource", map[string]interface{}{"type": "oceanbase_host", "id": state.ID.ValueString()})
	for _, id := range state.HostIDs {
		if err := r.client.DeleteHost(ctx, id); err != nil {
			var nfe *ocpclient.NotFoundError
			if errorAsNotFound(err, &nfe) {
				continue
			}
			resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete host %d", id), err.Error())
			return
		}
	}
}
