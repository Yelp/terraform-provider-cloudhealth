package cloudhealth

import (
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestMissingRefIds(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.Data(&terraform.InstanceState{
		ID: "1234",
		Attributes: map[string]string{
			"name":                                   "My Name",
			"include_in_reports":                     "true",
			"group.#":                                "2",
			"group.0.name":                           "New Group",
			"group.0.type":                           "filter",
			"group.0.rule.#":                         "1",
			"group.0.rule.0.asset":                   "AwsAccount",
			"group.0.rule.0.condition.#":             "1",
			"group.0.rule.0.condition.0.field.#":     "1",
			"group.0.rule.0.condition.0.field.0":     "Account Name",
			"group.0.rule.0.condition.0.op":          "=",
			"group.0.rule.0.condition.0.val":         "My Account",
			"group.1.name":                           "Existing Group",
			"group.1.type":                           "filter",
			"group.1.ref_id":                         "1",
			"group.1.rule.#":                         "1",
			"group.1.rule.0.asset":                   "AwsAsset",
			"group.1.rule.0.condition.#":             "1",
			"group.1.rule.0.condition.0.field.#":     "0",
			"group.1.rule.0.condition.0.tag_field.#": "1",
			"group.1.rule.0.condition.0.tag_field.0": "Name",
			"group.1.rule.0.condition.0.op":          "=",
			"group.1.rule.0.condition.0.val":         "My Name",
		},
	})
	b, err := tfToJson(rd)
	assert.Nil(t, err)

	newRD := resource.TestResourceData()
	jsonToTF(b, newRD)

	refId := rd.Get("group.0.ref_id")
	assert.NotEmpty(t, refId)
	assert.NotEqual(t, refId, "1")
}

func TestDynamicGroupsPreserved(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.Data(&terraform.InstanceState{
		ID: "1234",
		Attributes: map[string]string{
			"name":                 "My Name",
			"include_in_reports":   "true",
			"group.#":              "1",
			"group.0.name":         "New Group",
			"group.0.ref_id":       "1",
			"group.0.type":         "categorize",
			"group.0.rule.#":       "1",
			"group.0.rule.0.asset": "AwsAccount",
			"constant.#":           "2",

			"constant.0.constant_type": "Dynamic Group",
			"constant.0.ref_id":        "2",
			"constant.0.name":          "My Account",
			"constant.0.val":           "My Account",
			"constant.0.blk_id":        "1",

			"constant.1.constant_type": "Dynamic Group Block",
			"constant.1.ref_id":        "1",
			"constant.1.name":          "New Group",
		},
	})
	b, err := tfToJson(rd)
	assert.Nil(t, err)

	newRD := resource.TestResourceData()
	jsonToTF(b, newRD)

	assertEqual(t, newRD, "constant.#", 2)
	assertEqual(t, newRD, "constant.0.constant_type", "Dynamic Group")
	assertEqual(t, newRD, "constant.0.ref_id", "2")
	assertEqual(t, newRD, "constant.0.name", "My Account")
	assertEqual(t, newRD, "constant.0.val", "My Account")
	assertEqual(t, newRD, "constant.0.blk_id", "1")

	assertEqual(t, newRD, "constant.1.constant_type", "Dynamic Group Block")
	assertEqual(t, newRD, "constant.1.ref_id", "1")
	assertEqual(t, newRD, "constant.1.name", "New Group")
}
