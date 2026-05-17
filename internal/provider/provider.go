package provider

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oceanbase/terraform-provider-oceanbase/internal/ocpclient"
)

type oceanbaseProvider struct{}

func New() func() provider.Provider {
	return func() provider.Provider { return &oceanbaseProvider{} }
}

type providerModel struct {
	OcpURL          types.String `tfsdk:"ocp_url"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	TaskMode        types.String `tfsdk:"task_mode"`
	PollingTimeout  types.String `tfsdk:"polling_timeout"`
	PollingInterval types.String `tfsdk:"polling_interval"`
}

func (p *oceanbaseProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oceanbase"
	resp.Version = "0.1.0"
}

func (p *oceanbaseProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ocp_url":          schema.StringAttribute{Optional: true, Description: "OCP 服务地址，未设置时回退到 OCP_URL 环境变量"},
			"username":         schema.StringAttribute{Optional: true, Description: "OCP 用户名，未设置时回退到 OCP_USERNAME 环境变量"},
			"password":         schema.StringAttribute{Optional: true, Sensitive: true, Description: "OCP 密码，未设置时回退到 OCP_PASSWORD 环境变量"},
			"task_mode":        schema.StringAttribute{Optional: true, Description: "异步任务模式：polling（默认）或 fire_and_forget"},
			"polling_timeout":  schema.StringAttribute{Optional: true, Description: "异步任务最大等待时间，如 '30m'，默认 30m"},
			"polling_interval": schema.StringAttribute{Optional: true, Description: "任务状态轮询间隔，如 '10s'，默认 10s"},
		},
	}
}

func (p *oceanbaseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pick := func(v types.String, env string) string {
		if !v.IsNull() && v.ValueString() != "" {
			return v.ValueString()
		}
		return os.Getenv(env)
	}

	ocpURL := pick(cfg.OcpURL, "OCP_URL")
	username := pick(cfg.Username, "OCP_USERNAME")
	password := pick(cfg.Password, "OCP_PASSWORD")

	if ocpURL == "" || username == "" || password == "" {
		resp.Diagnostics.AddError("OCP 凭证缺失", "必须设置 ocp_url、username、password（可通过 HCL 或 OCP_URL/OCP_USERNAME/OCP_PASSWORD 环境变量）")
		return
	}

	client := ocpclient.NewClient(ocpURL, username, password)
	if !cfg.TaskMode.IsNull() && cfg.TaskMode.ValueString() != "" {
		client.TaskMode = ocpclient.TaskMode(cfg.TaskMode.ValueString())
	}
	if !cfg.PollingTimeout.IsNull() && cfg.PollingTimeout.ValueString() != "" {
		d, err := time.ParseDuration(cfg.PollingTimeout.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("polling_timeout 无效", err.Error())
			return
		}
		client.PollingTimeout = d
	}
	if !cfg.PollingInterval.IsNull() && cfg.PollingInterval.ValueString() != "" {
		d, err := time.ParseDuration(cfg.PollingInterval.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("polling_interval 无效", err.Error())
			return
		}
		client.PollingInterval = d
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *oceanbaseProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
		NewTenantResource,
		NewObproxyResource,
		NewArbitrationResource,
		NewHostResource,
	}
}

func (p *oceanbaseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewClustersDataSource,
		NewHostsDataSource,
	}
}
