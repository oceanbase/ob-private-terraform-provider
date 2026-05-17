package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type oceanbaseProvider struct{}

func New() func() provider.Provider {
	return func() provider.Provider { return &oceanbaseProvider{} }
}

func (p *oceanbaseProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oceanbase"
	resp.Version = "0.1.0"
}

func (p *oceanbaseProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *oceanbaseProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *oceanbaseProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *oceanbaseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
