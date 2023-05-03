package hashicups

import (
	"context"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	dreamhostAPIKeyEnvVarName = "DREAMHOST_API_KEY" // nolint:gosec
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc(dreamhostAPIKeyEnvVarName, nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"dreamhost_dns_record": resourceDNSRecord(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey, ok := d.Get("api_key").(string)
	if !ok {
		return nil, diag.Errorf("could not obtain api_key from configuration")
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if apiKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing Dreamhost API key",
			Detail: "Could not determine Dreamhost API key - both configuration and env var " +
				dreamhostAPIKeyEnvVarName + " are not set",
		})
		return nil, diags
	}

	api, err := dreamhostapi.NewClient(apiKey, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Dreamhost client",
			Detail:   "Unable to create Dreamhost client",
		})

		return nil, diags
	}

	return api, diags
}
