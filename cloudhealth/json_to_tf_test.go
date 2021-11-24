package cloudhealth

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"

	"encoding/json"
	"fmt"
	"io/ioutil"

	"testing"
)

func TestJsonToTFUnrecognizedData(t *testing.T) {
	// We should error if there is any field in the input we don't recognize
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()
	err := jsonToTF([]byte(`{"schema": {"name": "a", "include_in_reports": "true", "some_key": "some_value"}}`), rd)
	assert.NotNil(t, err)
}

func TestJsonToTFStatic(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()

	bytes, err := ioutil.ReadFile("../test/static_perspective.json")
	err = jsonToTF(bytes, rd)
	assert.Nil(t, err)

	assertEqual(t, rd, "name", "My Name")
	assertEqual(t, rd, "include_in_reports", true)
	assertEqual(t, rd, "group.#", 3)
	assertEqual(t, rd, "group.0.name", "Group One")
	assertEqual(t, rd, "group.0.type", "filter")
	assertEqual(t, rd, "group.0.ref_id", "1")
	assertEqual(t, rd, "group.0.rule.#", 1)
	assertEqual(t, rd, "group.0.rule.0.asset", "AwsAccount")
	assertEqual(t, rd, "group.0.rule.0.condition.#", 1)
	assertEqual(t, rd, "group.0.rule.0.condition.0.field.#", 1)
	assertEqual(t, rd, "group.0.rule.0.condition.0.field.0", "Account Name")
	assertEqual(t, rd, "group.0.rule.0.condition.0.op", "=")
	assertEqual(t, rd, "group.0.rule.0.condition.0.val", "My Account")

	assertEqual(t, rd, "group.1.name", "Group Two")
	assertEqual(t, rd, "group.1.ref_id", "2")
	assertEqual(t, rd, "group.1.type", "filter")
	assertEqual(t, rd, "group.1.rule.#", 1)
	assertEqual(t, rd, "group.1.rule.0.asset", "AwsAccount")
	assertEqual(t, rd, "group.1.rule.0.combine_with", "OR")
	assertEqual(t, rd, "group.1.rule.0.condition.#", 2)
	assertEqual(t, rd, "group.1.rule.0.condition.0.field.#", 1)
	assertEqual(t, rd, "group.1.rule.0.condition.0.field.0", "Account Name")
	assertEqual(t, rd, "group.1.rule.0.condition.0.op", "Contains")
	assertEqual(t, rd, "group.1.rule.0.condition.0.val", "Some Account")
	assertEqual(t, rd, "group.1.rule.0.condition.1.field.#", 1)
	assertEqual(t, rd, "group.1.rule.0.condition.1.field.0", "Account Name")
	assertEqual(t, rd, "group.1.rule.0.condition.1.op", "Contains")
	assertEqual(t, rd, "group.1.rule.0.condition.1.val", "Another Account")

	assertEqual(t, rd, "group.2.name", "Group Three")
	assertEqual(t, rd, "group.2.rule.#", 1)
	assertEqual(t, rd, "group.2.type", "filter")
	assertEqual(t, rd, "group.2.rule.0.asset", "AwsAsset")
	assertEqual(t, rd, "group.2.rule.0.condition.#", 1)
	assertEqual(t, rd, "group.2.rule.0.condition.0.field.#", 0)
	assertEqual(t, rd, "group.2.rule.0.condition.0.tag_field.#", 1)
	assertEqual(t, rd, "group.2.rule.0.condition.0.tag_field.0", "team")
	assertEqual(t, rd, "group.2.rule.0.condition.0.op", "=")
	assertEqual(t, rd, "group.2.rule.0.condition.0.val", "My Team")

	assertEqual(t, rd, "constant.#", 4)
	assertEqual(t, rd, "constant.0.constant_type", "Static Group")
	assertEqual(t, rd, "constant.0.name", "Group One")
	assertEqual(t, rd, "constant.0.ref_id", "1")

	assertEqual(t, rd, "constant.2.constant_type", "Static Group")
	assertEqual(t, rd, "constant.1.name", "Group Two")
	assertEqual(t, rd, "constant.1.ref_id", "2")

	assertEqual(t, rd, "constant.2.constant_type", "Static Group")
	assertEqual(t, rd, "constant.2.name", "Group Three")
	assertEqual(t, rd, "constant.2.ref_id", "3")

	assertEqual(t, rd, "constant.3.constant_type", "Static Group")
	assertEqual(t, rd, "constant.3.name", "Other")
	assertEqual(t, rd, "constant.3.ref_id", "4")
	assertEqual(t, rd, "constant.3.is_other", "true")
}

func TestJsonToTFToJsonDynamic(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()

	originalBytes, err := ioutil.ReadFile("../test/dynamic_perspective.json")
	err = jsonToTF(originalBytes, rd)
	assert.Nil(t, err)
	assertEqual(t, rd, "include_in_reports", false)

	resultBytes, err := tfToJson(rd)
	assert.Nil(t, err)
	assertJsonEqual(t, originalBytes, resultBytes)
}

func TestJsonToTFToJsonStatic(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()

	originalBytes, err := ioutil.ReadFile("../test/static_perspective.json")
	err = jsonToTF(originalBytes, rd)
	assert.Nil(t, err)
	assertEqual(t, rd, "include_in_reports", true)

	resultBytes, err := tfToJson(rd)
	assert.Nil(t, err)
	assertJsonEqual(t, originalBytes, resultBytes)
}

func TestJsonToTFDynamic(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()

	bytes, err := ioutil.ReadFile("../test/dynamic_perspective.json")
	err = jsonToTF(bytes, rd)
	assert.Nil(t, err)

	assertEqual(t, rd, "name", "My Dynamic")
	assert.False(t, rd.Get("include_in_reports").(bool), "include_in_reports")
	assertEqual(t, rd, "group.#", 2)
	assertEqual(t, rd, "group.0.name", "Group One")
	assertEqual(t, rd, "group.0.ref_id", "1")
	assertEqual(t, rd, "group.0.type", "categorize")
	assertEqual(t, rd, "group.0.rule.#", 1)
	assertEqual(t, rd, "group.0.rule.0.asset", "AwsAsset")
	assertEqual(t, rd, "group.0.rule.0.tag_field.#", 1)
	assertEqual(t, rd, "group.0.rule.0.tag_field.0", "my_tag")
	assertEqual(t, rd, "group.0.rule.0.condition.#", 1)
	assertEqual(t, rd, "group.0.rule.0.condition.0.field.#", 1)
	assertEqual(t, rd, "group.0.rule.0.condition.0.field.0", "Account Name")
	assertEqual(t, rd, "group.0.rule.0.condition.0.op", "!=")
	assertEqual(t, rd, "group.0.rule.0.condition.0.val", "Excluded Account")
	assertEqual(t, rd, "group.1.name", "Group Two")
	assertEqual(t, rd, "group.1.ref_id", "2")
	assertEqual(t, rd, "group.1.type", "categorize")
	assertEqual(t, rd, "group.1.rule.#", 1)
	assertEqual(t, rd, "group.1.rule.0.asset", "AwsRedshiftCluster")
	assertEqual(t, rd, "group.1.rule.0.field.#", 1)
	assertEqual(t, rd, "group.1.rule.0.field.0", "Cluster Identifier")
	assertEqual(t, rd, "group.1.rule.0.condition.#", 0)
	assertEqual(t, rd, "constant.#", 7)

	// These are populated in exactly the order that they're seen in the JSON
	assertEqual(t, rd, "constant.0.constant_type", "Static Group")
	assertEqual(t, rd, "constant.0.ref_id", "7")
	assertEqual(t, rd, "constant.0.name", "Other")
	assertEqual(t, rd, "constant.0.is_other", "true")

	assertEqual(t, rd, "constant.1.constant_type", "Dynamic Group")
	assertEqual(t, rd, "constant.1.ref_id", "5")
	assertEqual(t, rd, "constant.1.blk_id", "1")
	assertEqual(t, rd, "constant.1.name", "ValC")
	assertEqual(t, rd, "constant.1.val", "ValC")

	assertEqual(t, rd, "constant.2.constant_type", "Dynamic Group")
	assertEqual(t, rd, "constant.2.ref_id", "3")
	assertEqual(t, rd, "constant.2.blk_id", "2")
	assertEqual(t, rd, "constant.2.name", "ValA")
	assertEqual(t, rd, "constant.2.val", "ValA")

	assertEqual(t, rd, "constant.3.constant_type", "Dynamic Group")
	assertEqual(t, rd, "constant.3.ref_id", "4")
	assertEqual(t, rd, "constant.3.blk_id", "2")
	assertEqual(t, rd, "constant.3.name", "ValB")
	assertEqual(t, rd, "constant.3.val", "ValB")

	assertEqual(t, rd, "constant.4.constant_type", "Dynamic Group")
	assertEqual(t, rd, "constant.4.ref_id", "6")
	assertEqual(t, rd, "constant.4.name", "Remaining")
	assertEqual(t, rd, "constant.4.val", "Remaining")

	assertEqual(t, rd, "constant.5.constant_type", "Dynamic Group Block")
	assertEqual(t, rd, "constant.5.ref_id", "1")
	assertEqual(t, rd, "constant.5.name", "Group One")

	assertEqual(t, rd, "constant.6.constant_type", "Dynamic Group Block")
	assertEqual(t, rd, "constant.6.ref_id", "2")
	assertEqual(t, rd, "constant.6.name", "Group Two")
}

func TestReorderGroup(t *testing.T) {
	// Ensure that groups keep their ref_id values if they're reordered in config

	// Step 0: Load perspective from JSON into a resource.Data
	// Step 2: Run plan on a new version of that config where groups are re-ordered. Verify refIds

	// Load perspective from JSON into a resource.Data
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()
	originalBytes, err := ioutil.ReadFile("../test/static_perspective.json")
	err = jsonToTF(originalBytes, rd)
	assert.Nil(t, err)

	// Verify ref_ids
	assertEqual(t, rd, "group.0.ref_id", "1")
	assertEqual(t, rd, "group.1.ref_id", "2")
	assertEqual(t, rd, "group.2.ref_id", "3")

	// Simulate re-arranging groups via config - move 2 before 1
	groups := rd.Get("group").([]interface{})
	newGroups := []map[string]interface{}{
		groups[1].(map[string]interface{}),
		groups[0].(map[string]interface{}),
		groups[2].(map[string]interface{}),
	}

	// Simulate the "breakage" of group.ref_id: they do not follow the list reordering, so they are still 1,2,3
	newGroups[0]["ref_id"] = "1"
	newGroups[1]["ref_id"] = "2"
	newGroups[2]["ref_id"] = "3"
	err = rd.Set("group", newGroups)
	assert.Nil(t, err)

	// Convert to json and back to TF
	b, err := tfToJson(rd)
	assert.Nil(t, err)
	newRD := resource.TestResourceData()
	jsonToTF(b, newRD)

	// The group ref_ids should now be correct!
	assertEqual(t, newRD, "group.0.ref_id", "2")
	assertEqual(t, newRD, "group.1.ref_id", "1")
	assertEqual(t, newRD, "group.2.ref_id", "3")
}

func TestRenameGroup(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()
	originalBytes, err := ioutil.ReadFile("../test/static_perspective.json")
	err = jsonToTF(originalBytes, rd)
	assert.Nil(t, err)

	// Give second group a new name
	groups := rd.Get("group").([]interface{})
	groups[1].(map[string]interface{})["name"] = "My New Name"
	err = rd.Set("group", groups)
	assert.Nil(t, err)

	// Convert to json and back to TF
	b, err := tfToJson(rd)
	assert.Nil(t, err)
	newRD := resource.TestResourceData()
	jsonToTF(b, newRD)

	assertEqual(t, newRD, "group.0.ref_id", "1")
	assertEqual(t, newRD, "group.0.name", "Group One")
	assertEqual(t, newRD, "group.1.ref_id", "2")
	assertEqual(t, newRD, "group.1.name", "My New Name")
	assertEqual(t, newRD, "group.2.ref_id", "3")
	assertEqual(t, newRD, "group.2.name", "Group Three")
}

func TestRenameAndReorderGroup(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()
	originalBytes, err := ioutil.ReadFile("../test/static_perspective.json")
	err = jsonToTF(originalBytes, rd)
	assert.Nil(t, err)

	// Give second group a new name
	groups := rd.Get("group").([]interface{})
	groups[1].(map[string]interface{})["name"] = "My New Name"
	err = rd.Set("group", groups)
	assert.Nil(t, err)

	// Simulate re-arranging groups via config - move 2 before 1
	newGroups := []map[string]interface{}{
		groups[1].(map[string]interface{}),
		groups[0].(map[string]interface{}),
		groups[2].(map[string]interface{}),
	}

	// Simulate the "breakage" of group.ref_id: they do not follow the list reordering, so they are still 1,2,3
	newGroups[0]["ref_id"] = "1"
	newGroups[1]["ref_id"] = "2"
	newGroups[2]["ref_id"] = "3"

	err = rd.Set("group", newGroups)
	assert.Nil(t, err)

	// Convert to json and back to TF
	b, err := tfToJson(rd)
	assert.Nil(t, err)
	newRD := resource.TestResourceData()
	jsonToTF(b, newRD)

	// The first ref_id should be a higher int than the rest of the items in constants
	refId := rd.Get("group.0.ref_id").(string)
	assert.NotEqual(t, refId, "2")
	assert.Equal(t, "5", refId)

	assertEqual(t, newRD, "group.0.name", "My New Name")
	assertEqual(t, newRD, "group.1.ref_id", "1")
	assertEqual(t, newRD, "group.1.name", "Group One")
	assertEqual(t, newRD, "group.2.ref_id", "3")
	assertEqual(t, newRD, "group.2.name", "Group Three")
}

func TestRemoveDynamicGroupBlock(t *testing.T) {
	resource := resourceCHTPerspective()
	rd := resource.TestResourceData()
	originalBytes, err := ioutil.ReadFile("../test/dynamic_perspective.json")
	err = jsonToTF(originalBytes, rd)
	assert.Nil(t, err)

	// Remove Group One
	groups := rd.Get("group").([]interface{})
	newGroups := []map[string]interface{}{
		groups[1].(map[string]interface{}),
	}

	// Simulate the "breakage" of group.ref_id: they do not follow the list
	// reordering, so the second-now-first group has ref_id 1
	newGroups[0]["ref_id"] = "2"
	err = rd.Set("group", newGroups)
	assert.Nil(t, err)

	// Convert to json and back to TF
	b, err := tfToJson(rd)
	assert.Nil(t, err)
	newRD := resource.TestResourceData()
	jsonToTF(b, newRD)

	// The group ref_ids should now be correct
	assertEqual(t, newRD, "group.0.ref_id", "2")

	// We should lose two constants: one for the Dynamic Group Block, and one
	// for the Dynamic Group inside it
	assertEqual(t, newRD, "constant.#", 5)
}

func assertEqual(t *testing.T, rd *schema.ResourceData, field string, expected interface{}) {
	actual := rd.Get(field)
	assert.Equal(t, expected, actual, field)
}

func assertJsonEqual(t *testing.T, expected []byte, actual []byte) {
	differ := gojsondiff.New()
	diff, err := differ.Compare(expected, actual)
	assert.Nil(t, err)
	if diff.Modified() {
		fmt.Println("JSON differs")
		t.Fail()
		printJsonDiff(diff, expected)
		fmt.Println("expected", string(expected))
		fmt.Println("actual", string(actual))
	}
}

func printJsonDiff(diff gojsondiff.Diff, b []byte) {
	var j map[string]interface{}
	json.Unmarshal(b, &j)

	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
		Coloring:       true,
	}

	f := formatter.NewAsciiFormatter(j, config)
	diffString, err := f.Format(diff)
	if err != nil {
		// No error can occur
	}
	fmt.Print(diffString)

	formatter2 := formatter.NewDeltaFormatter()
	diffString, _ = formatter2.Format(diff)
	fmt.Print(diffString)
}
