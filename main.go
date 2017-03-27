package main

import (
	"github.com/Yelp/terraform-provider-cloudhealth/cloudhealth"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudhealth.Provider,
	})
}
