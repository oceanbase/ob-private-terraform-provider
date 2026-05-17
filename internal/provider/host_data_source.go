package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oceanbase/terraform-provider-oceanbase/internal/ocpclient"
)

type hostsDataSource struct{ client *ocpclient.Client }

func NewHostsDataSource() datasource.DataSource { return &hostsDataSource{} }

type hostsDataSourceModel struct {
	ID     types.String   `tfsdk:"id"`
	Status types.String   `tfsdk:"status"`
	IdcID  types.Int64    `tfsdk:"idc_id"`
	Hosts  []hostListItem `tfsdk:"hosts"`
}

type hostListItem struct {
	ID             types.String `tfsdk:"id"`
	InnerIpAddress types.String `tfsdk:"inner_ip_address"`
	Status         types.String `tfsdk:"status"`
	IdcID          types.Int64  `tfsdk:"idc_id"`
	TypeID         types.Int64  `tfsdk:"type_id"`
	Alias          types.String `tfsdk:"alias"`
}

func (d *hostsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosts"
}

func (d *hostsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true},
			"status": schema.StringAttribute{Optional: true, Description: "按主机状态过滤，如 AVAILABLE"},
			"idc_id": schema.Int64Attribute{Optional: true, Description: "按 IDC ID 过滤"},
			"hosts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"inner_ip_address": schema.StringAttribute{Computed: true},
						"status":           schema.StringAttribute{Computed: true},
						"idc_id":           schema.Int64Attribute{Computed: true},
						"type_id":          schema.Int64Attribute{Computed: true},
						"alias":            schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *hostsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*ocpclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Provider 数据类型异常", fmt.Sprintf("实际收到 %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *hostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg hostsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	filter := &ocpclient.HostListFilter{Status: cfg.Status.ValueString()}
	if !cfg.IdcID.IsNull() {
		v := cfg.IdcID.ValueInt64()
		filter.IdcID = &v
	}
	lst, err := d.client.ListHosts(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError("查询主机列表失败", err.Error())
		return
	}
	out := hostsDataSourceModel{ID: types.StringValue("hosts"), Status: cfg.Status, IdcID: cfg.IdcID}
	for _, h := range lst {
		out.Hosts = append(out.Hosts, hostListItem{
			ID:             types.StringValue(strconv.FormatInt(h.ID, 10)),
			InnerIpAddress: types.StringValue(h.InnerIpAddress),
			Status:         types.StringValue(h.Status),
			IdcID:          types.Int64Value(h.IdcID),
			TypeID:         types.Int64Value(h.TypeID),
			Alias:          types.StringValue(h.Alias),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
