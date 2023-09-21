package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &DetectifyProvider{}

// DetectifyProvider defines the provider implementation.
type DetectifyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// DetectifyProviderModel describes the provider data model.
type DetectifyProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api:key"`
}

func (p *DetectifyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "detectify"
	resp.Version = p.version
}

func (p *DetectifyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Detectify API key",
				Required:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
		},
	}
}

func (p *DetectifyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DetectifyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DetectifyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAssetResource,
	}
}

func (p *DetectifyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAssetDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DetectifyProvider{
			version: version,
		}
	}
}
