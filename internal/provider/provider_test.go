package provider_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/jsvensson/terraform-provider-detectify/internal/provider"
	"github.com/stretchr/testify/require"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"detectify": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testProviderPreCheck(t *testing.T) {
	// TODO: Validate provider setup
}

func TestCalculateHMACSignature(t *testing.T) {
	// path := "/v2/domains/"
	var ts int64 = 1519829567
	apiKey := "10840b0f938942feafb7186de74b9682"
	secretKey := "0vyTnawJRFn0Q9tWLTM188Olizc72JczHSXoIlsPQIc="

	req, err := http.NewRequest(http.MethodGet, "http://localhost/v2/domains", nil)
	require.NoError(t, err)

	expected := "6jpu6S4cQwEY4uLk+xELSe1RhajVJP0QEDpGWZ5T+U0="
	actual := provider.CalculateSignature(req, apiKey, secretKey, time.Unix(ts, 0))

	require.Equal(t, expected, actual)
}
