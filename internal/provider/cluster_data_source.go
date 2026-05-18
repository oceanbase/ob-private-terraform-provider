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

type clustersDataSource struct{ client *ocpclient.Client }

func NewClustersDataSource() datasource.DataSource { return &clustersDataSource{} }

type clustersDataSourceModel struct {
	ID       types.String      `tfsdk:"id"`
	Clusters []clusterListItem `tfsdk:"clusters"`
}

type clusterListItem struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Status      types.String `tfsdk:"status"`
	Type        types.String `tfsdk:"type"`
	FullVersion types.String `tfsdk:"full_version"`
}

func (d *clustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ob_clusters"
}

func (d *clustersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"clusters": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"status":       schema.StringAttribute{Computed: true},
						"type":         schema.StringAttribute{Computed: true},
						"full_version": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *clustersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*ocpclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *ocpclient.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *clustersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	list, err := d.client.ListClusters(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list clusters", err.Error())
		return
	}
	state := clustersDataSourceModel{ID: types.StringValue("clusters")}
	for _, c := range list {
		state.Clusters = append(state.Clusters, clusterListItem{
			ID:          types.StringValue(strconv.FormatInt(c.ID, 10)),
			Name:        types.StringValue(c.Name),
			Status:      types.StringValue(c.Status),
			Type:        types.StringValue(c.Type),
			FullVersion: types.StringValue(c.FullVersion),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
