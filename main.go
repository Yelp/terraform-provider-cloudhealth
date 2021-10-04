package main

import (
	"cloudhealth/cloudhealth"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudhealth.Provider,
	})
}
