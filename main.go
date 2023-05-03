package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/adamantal/terraform-provider-dreamhost/dreamhost"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: dreamhost.Provider,
	})
}
