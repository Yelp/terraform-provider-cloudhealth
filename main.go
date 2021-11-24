package main

import (
	"cloudhealth/cloudhealth"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudhealth.Provider,
	})
}
