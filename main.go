package main

import (
	"github.com/bchess/terraform-provider-cloudhealth/cloudhealth"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudhealth.Provider,
	})
}
