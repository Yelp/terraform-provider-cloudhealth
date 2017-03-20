package cloudhealth

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
)

func tfToJson(d *schema.ResourceData) (rawData []byte, err error) {
	var pj PerspectiveJSON

	constants := []*ConstantJSON{
		NewConstantJSON(StaticGroupType),
		NewConstantJSON(DynamicGroupType),
		NewConstantJSON(DynamicGroupBlockType),
	}

	constantsByType := make(map[string]*ConstantJSON)
	for _, constant := range constants {
		constantsByType[constant.Type] = constant
	}

	name, ok := d.GetOk("name")
	if !ok {
		return nil, fmt.Errorf("Required name field")
	}
	pj.Schema.Name = name.(string)

	includeInReports := d.Get("include_in_reports")
	pj.Schema.Include_in_reports = strconv.FormatBool(includeInReports.(bool))

	tfGroups := getArray(d, "group")
	otherGroups := getArray(d, "other_group")

	if len(tfGroups) > 0 {
		fillInMissingRefIDs(tfGroups, otherGroups)
		err = d.Set("group", tfGroups)
		if err != nil {
			return nil, err
		}
	}

	for _, tfGroup := range tfGroups {
		tfGroup := tfGroup.(map[string]interface{})
		refId := tfGroup["ref_id"].(string)

		var constantType string
		if tfGroup["type"].(string) == "categorize" {
			// Convert any dynamic groups for this group (if it's a Dynamic Group Block)
			dynamicGroupConstantItems := dynamicGroupConstantItemsToJson(refId, tfGroup["dynamic_group"].([]interface{}))
			constantsByType[DynamicGroupType].List = append(constantsByType[DynamicGroupType].List, dynamicGroupConstantItems...)
			constantType = DynamicGroupBlockType
		} else if tfGroup["type"].(string) == "filter" {
			constantType = StaticGroupType
		} else {
			return nil, fmt.Errorf("Unknown group type: %s. Expected filter or categorize", tfGroup["type"])
		}

		// Convert any rules
		rules, err := rulesToJson(refId, constantType, tfGroup["rule"].([]interface{}))
		if err != nil {
			return nil, err
		}
		pj.Schema.Rules = append(pj.Schema.Rules, rules...)

		// Add a constant for this group
		constantItem := ConstantItem{
			Name:   tfGroup["name"].(string),
			Ref_id: tfGroup["ref_id"].(string),
		}
		constant := constantsByType[constantType]
		constant.List = append(constant.List, constantItem)
	}

	// Add constants for all other_group entries
	for _, otherGroup := range otherGroups {
		otherGroup := otherGroup.(map[string]interface{})
		constantType, constantItem := otherGroupToJson(otherGroup)

		constant := constantsByType[constantType]
		if constant == nil {
			return nil, fmt.Errorf("Unknown constant type %s", constantType)
		}
		constant.List = append(constant.List, constantItem)
	}

	// Only add constants that have something in them
	for _, constantGroup := range constants {
		if len(constantGroup.List) > 0 {
			pj.Schema.Constants = append(pj.Schema.Constants, *constantGroup)
		}
	}
	pj.Schema.Merges = make([]interface{}, 0)

	return json.MarshalIndent(pj, "", "  ")
}

func fillInMissingRefIDs(groups []interface{}, otherGroups []interface{}) {
	for _, g := range groups {
		g := g.(map[string]interface{})
		if g["ref_id"].(string) == "" {
			g["ref_id"], _ = uuid.GenerateUUID()
		}
	}
}

func dynamicGroupConstantItemsToJson(groupRefId string, dynamicGroups []interface{}) []ConstantItem {
	result := make([]ConstantItem, len(dynamicGroups))

	for idx, dg := range dynamicGroups {
		dg := dg.(map[string]interface{})
		blk_id := groupRefId
		result[idx] = ConstantItem{
			Name:   dg["name"].(string),
			Ref_id: dg["ref_id"].(string),
			Blk_id: &blk_id,
			Val:    dg["val"].(string),
		}
	}
	return result
}

func rulesToJson(groupRefId string, constantType string, rules []interface{}) (result []RuleJSON, err error) {
	result = make([]RuleJSON, len(rules))

	for ruleIdx, r := range rules {
		r := r.(map[string]interface{})

		rj := &result[ruleIdx]

		if constantType == DynamicGroupBlockType {
			rj.Ref_id = groupRefId
			rj.Type = "categorize"
		} else if constantType == StaticGroupType {
			rj.To = groupRefId
			rj.Type = "filter"
		} else {
			return nil, fmt.Errorf("Unrecognized group type %s", constantType)
		}

		rj.Asset = stringOrNil(r["asset"])
		rj.Field = convertStringArray(r["field"])
		rj.Tag_field = convertStringArray(r["tag_field"])

		if r["condition"] != nil {
			rj.Condition = conditionsToJson(r["condition"].([]interface{}))
		} else {
			rj.Condition = nil
		}
	}
	return result, nil
}

func conditionsToJson(conditions []interface{}) (result *ConditionJSON) {
	if len(conditions) == 0 {
		return nil
	}
	result = new(ConditionJSON)
	result.Clauses = make([]ClauseJSON, len(conditions))
	for idx, condition := range conditions {
		condition := condition.(map[string]interface{})
		result.Clauses[idx] = ClauseJSON{
			Field:     convertStringArray(condition["field"]),
			Tag_field: convertStringArray(condition["tag_field"]),
			Op:        stringOrNil(condition["op"]),
			Val:       stringOrNil(condition["val"]),
		}
	}
	return result
}

func otherGroupToJson(otherGroup map[string]interface{}) (constantType string, constantItem ConstantItem) {
	constantType = otherGroup["constant_type"].(string)
	constantItem = ConstantItem{
		Ref_id: stringOrNil(otherGroup["ref_id"]),
		Name:   stringOrNil(otherGroup["name"]),
		Val:    stringOrNil(otherGroup["val"]),
	}
	if constantType == DynamicGroupType {
		blk_id := stringOrNil(otherGroup["blk_id"])
		constantItem.Blk_id = &blk_id
	}
	if stringOrNil(otherGroup["is_other"]) == "true" {
		constantItem.Is_other = "true"
	}
	return constantType, constantItem
}

func convertStringArray(maybeStringArray interface{}) []string {
	if maybeStringArray == nil {
		return nil
	}
	ss := maybeStringArray.([]interface{})
	result := make([]string, len(ss))
	for idx, s := range ss {
		result[idx] = s.(string)
	}
	return result
}

func stringOrNil(s interface{}) string {
	if s == nil {
		return ""
	}
	return s.(string)
}

func getArray(d *schema.ResourceData, field string) []interface{} {
	if v, ok := d.GetOk(field); ok {
		return v.([]interface{})
	} else {
		return make([]interface{}, 0)
	}
}
