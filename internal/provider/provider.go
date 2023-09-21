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

// Ensure DetectifyProvider satisfies various provider interfaces.
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
	APIKey    types.String `tfsdk:"api_key"`
	Signature types.String `tfsdk:"signature"`
}

// DetectifyProviderData is used by resources and datasources to complete requests.
type DetectifyProviderData struct {
	Client    *http.Client
	Signature string
}

func (p *DetectifyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "detectify"
	resp.Version = p.version
}

func (p *DetectifyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Detectify API key.",
				Required:            true,
			},
			"signature": schema.StringAttribute{
				MarkdownDescription: "Signature for HMAC authentication. See [API documentation](https://developer.detectify.com/#section/Detectify-API/Authentication) for more information.",
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

	// add authentication headers
	headers := http.Header{}
	headers.Set("X-Detectify-Key", data.APIKey.ValueString())

	// wrap transport for client
	client := http.DefaultClient
	client.Transport = &transport{
		Transport: http.DefaultTransport,
		Headers:   headers,
	}

	providerData := DetectifyProviderData{
		Client:    client,
		Signature: data.Signature.ValueString(),
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
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

// custom transport with API credentials in headers
type transport struct {
	Transport http.RoundTripper
	Headers   http.Header
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, values := range t.Headers {
		req.Header[k] = values
	}

	return t.Transport.RoundTrip(req)
}
