package cloudhealth

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"cloudhealth": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
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
