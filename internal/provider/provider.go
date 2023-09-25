package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	APIKey types.String `tfsdk:"api_key"`
	Secret types.String `tfsdk:"secret"`
}

// DetectifyProviderData is used by resources and datasources to complete requests.
type DetectifyProviderData struct {
	Client *http.Client
	Secret string
}

func (p *DetectifyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "detectify"
	resp.Version = p.version
}

func (p *DetectifyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Detectify API key. May also be provided via `DETECTIFY_API_KEY` environment variable.",
				Required:            true,
				Sensitive:           true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "Secret used for HMAC signature. May also be provided via `DETECTIFY_SECRET` environment variable. " +
					"See [API documentation](https://developer.detectify.com/#section/Detectify-API/Authentication) for more information.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *DetectifyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Detectify provider")

	var config DetectifyProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Detectify API key",
			"", // TODO: error message that makes sense
		)
	}

	if config.Secret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Unknown Detectify secret",
			"", // TODO: error message that makes sense
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Read configuration default values from environment,
	// overriding with Terraform configuration values if set.
	apiKey := os.Getenv("DETECTIFY_API_KEY")
	secret := os.Getenv("DETECTIFY_SECRET")

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.Secret.IsNull() {
		secret = config.Secret.ValueString()
	}

	// If any expected configuration is missing, add errors with instructions.
	if len(apiKey) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Detectify API key",
			"The provider cannot create the Detectify API client as there is a missing or empty value for the Detectify API key. "+
				"Set the API key value in the configuration or use the DETECTIFY_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "api_key")
	ctx = tflog.SetField(ctx, "secret", secret)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "secret")

	// add authentication headers
	headers := http.Header{}
	headers.Set("X-Detectify-Key", apiKey)

	// wrap transport for client
	client := http.DefaultClient
	client.Transport = &transport{
		Transport: http.DefaultTransport,
		Headers:   headers,
		signature: config.Secret.ValueString(),
	}

	providerData := DetectifyProviderData{
		Client: client,
		Secret: config.Secret.ValueString(),
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData

	tflog.Debug(ctx, "Configured Detectify provider", map[string]any{"success": true})
}

// Resources defines the resources implemented in the provider.
func (p *DetectifyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAssetResource,
	}
}

// DataSources defines the data sources implemented in the provider.
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
	apiKey    string
	secret    string
	signature string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.signature) > 0 {
		ts := time.Now()
		signature := CalculateSignature(req, t.apiKey, t.secret, ts)

		t.Headers.Set("X-Detectify-Timestamp", strconv.FormatInt(ts.Unix(), 10))
		t.Headers.Set("X-Detectify-Signature", signature)
	}

	for k, values := range t.Headers {
		req.Header[k] = values
	}

	return t.Transport.RoundTrip(req)
}

// Calculate the HMAC signature for the request.
func CalculateSignature(req *http.Request, apiKey, secretKey string, timestamp time.Time) string {
	key, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		panic(err)
	}

	// TODO: Issue with reading body like this?

	value := fmt.Sprintf("%s;%s;%s;%d;%s", req.Method, req.URL.Path, apiKey, timestamp.Unix(), req.Body)
	fmt.Println(value)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
