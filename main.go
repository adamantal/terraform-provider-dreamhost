package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/adamantal/terraform-provider-dreamhost/hashicups"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: hashicups.Provider,
	})
}
