package cloudhealth

import (
	"github.com/hashicorp/terraform/helper/schema"
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
	assertEqual(t, rd, "group.#", 2)
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
	assertEqual(t, rd, "group.1.rule.#", 1)
	assertEqual(t, rd, "group.1.rule.0.asset", "AwsAccount")
	assertEqual(t, rd, "group.1.rule.0.condition.#", 1)
	assertEqual(t, rd, "group.1.rule.0.condition.0.field.#", 1)
	assertEqual(t, rd, "group.1.rule.0.condition.0.field.0", "Account Name")
	assertEqual(t, rd, "group.1.rule.0.condition.0.op", "Contains")
	assertEqual(t, rd, "group.1.rule.0.condition.0.val", "Another")
	assertEqual(t, rd, "other_group.#", 1)
	assertEqual(t, rd, "other_group.0.name", "Other")
	assertEqual(t, rd, "other_group.0.constant_type", "Static Group")
	assertEqual(t, rd, "other_group.0.ref_id", "3")
	assertEqual(t, rd, "other_group.0.is_other", "true")
}

func TestJsonToTFToJson(t *testing.T) {
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
	assertEqual(t, rd, "group.0.dynamic_group.#", 1)
	assertEqual(t, rd, "group.0.dynamic_group.0.ref_id", "5")
	assertEqual(t, rd, "group.0.dynamic_group.0.name", "ValC")
	assertEqual(t, rd, "group.0.dynamic_group.0.val", "ValC")
	assertEqual(t, rd, "group.1.name", "Group Two")
	assertEqual(t, rd, "group.1.ref_id", "2")
	assertEqual(t, rd, "group.1.type", "categorize")
	assertEqual(t, rd, "group.1.rule.#", 1)
	assertEqual(t, rd, "group.1.rule.0.asset", "AwsRedshiftCluster")
	assertEqual(t, rd, "group.1.rule.0.field.#", 1)
	assertEqual(t, rd, "group.1.rule.0.field.0", "Cluster Identifier")
	assertEqual(t, rd, "group.1.rule.0.condition.#", 0)
	assertEqual(t, rd, "group.1.dynamic_group.#", 2)
	assertEqual(t, rd, "group.1.dynamic_group.0.ref_id", "3")
	assertEqual(t, rd, "group.1.dynamic_group.0.name", "ValA")
	assertEqual(t, rd, "group.1.dynamic_group.0.val", "ValA")
	assertEqual(t, rd, "group.1.dynamic_group.1.ref_id", "4")
	assertEqual(t, rd, "group.1.dynamic_group.1.name", "ValB")
	assertEqual(t, rd, "group.1.dynamic_group.1.val", "ValB")
	assertEqual(t, rd, "other_group.#", 2)
	assertEqual(t, rd, "other_group.0.ref_id", "7")
	assertEqual(t, rd, "other_group.0.constant_type", "Static Group")
	assertEqual(t, rd, "other_group.0.name", "Other")
	assertEqual(t, rd, "other_group.0.is_other", "true")
	assertEqual(t, rd, "other_group.1.ref_id", "6")
	assertEqual(t, rd, "other_group.1.constant_type", "Dynamic Group")
	assertEqual(t, rd, "other_group.1.name", "Remaining")
	assertEqual(t, rd, "other_group.1.val", "Remaining")
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
