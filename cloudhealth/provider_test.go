package cloudhealth

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"cloudhealth": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

const testAccCreateConfig = `
resource "cloudhealth_perspective" "acc_test_owner_tag" {
  name               = "acc test owner tag"
  include_in_reports = false
	hard_delete        = true

  group {
    name = "OwnerAccTest"
    type = "categorize"

    rule {
      asset     = "AwsAsset"
      tag_field = ["owner"]
    }
  }
}
`

func TestAccCheckCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCreateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudhealth_perspective.acc_test_owner_tag", "name", "acc test owner tag"),
				),
			},
		},
	})
}
