package cloudhealth

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ChtMeta struct {
	apiKey string
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CHT_API_KEY", nil),
				Description: "API key for Cloudhealth",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"cloudhealth_perspective": resourceCHTPerspective(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	key := ""
	if k, ok := d.GetOk("key"); ok {
		key = k.(string)
	} else {
		return nil, diag.FromErr(errors.New("Must set CHT_API_KEY or provide a 'key' to the provider"))
	}
	meta := ChtMeta{
		apiKey: key,
	}
	return &meta, nil
}
