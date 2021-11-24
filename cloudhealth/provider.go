package cloudhealth

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type ChtMeta struct {
	apiKey string
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CHT_API_KEY", ""),
				Description: "API key for Cloudhealth",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"cloudhealth_perspective": resourceCHTPerspective(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	key := d.Get("key").(string)
	if key == "" {
		return nil, errors.New("Must set CHT_API_KEY or provide a 'key' to the provider")
	}
	meta := ChtMeta{
		apiKey: key,
	}
	return &meta, nil
}
