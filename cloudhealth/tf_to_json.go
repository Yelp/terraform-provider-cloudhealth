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
	tfConstants := getArray(d, "constant")

	if len(tfGroups) > 0 {
		fixRefIDs(tfGroups, tfConstants)
		err = d.Set("group", tfGroups)
		if err != nil {
			return nil, err
		}
	}

	for _, tfGroup := range tfGroups {
		tfGroup := tfGroup.(map[string]interface{})
		refId := tfGroup["ref_id"].(string)
		name := tfGroup["name"].(string)
		groupType := tfGroup["type"].(string)

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
		rules, err := rulesToJson(refId, name, groupType, tfGroup["rule"].([]interface{}))
		if err != nil {
			return nil, err
		}
		pj.Schema.Rules = append(pj.Schema.Rules, rules...)

		// Add a constant for this group
		constantItem := ConstantItem{
			Name:   name,
			Ref_id: refId,
		}
		constant := constantsByType[constantType]
		constant.List = append(constant.List, constantItem)
	}

	err = addOtherConstants(tfConstants, constantsByType)
	if err != nil {
		return nil, err
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

func fixRefIDs(groups []interface{}, constants []interface{}) {
	// Goal of this is to reconcile the ref_id on groups with the ones in constants
	refIdByNameFromBindings := make(map[string]string)
	for _, c := range constants {
		c := c.(map[string]interface{})
		refIdByNameFromBindings[c["name"].(string)] = c["ref_id"].(string)
	}
	usedRefIds := make(map[string]bool)

	// Go through and apply the ref_id from the binding to anything that matches the same name in the group
	for _, g := range groups {
		g := g.(map[string]interface{})
		groupName := g["name"].(string)
		if bindingRefId, ok := refIdByNameFromBindings[groupName]; ok {
			g["ref_id"] = bindingRefId
			usedRefIds[bindingRefId] = true
		}
	}

	// Now for any group who name is not in bindings, either use its exising
	// ref_id (we assume this meant a rename) or, if it doesn't have one,
	// generate a unique one
	for _, g := range groups {
		g := g.(map[string]interface{})
		groupName := g["name"].(string)
		groupRefId := g["ref_id"].(string)

		_, inBindings := refIdByNameFromBindings[groupName]
		if inBindings {
			continue
		}
		if groupRefId != "" && usedRefIds[groupRefId] == false {
			// Group was renamed; stick with the existing groupRefId
			continue
		}
		// Group is new - assign a new ref id
		g["ref_id"], _ = uuid.GenerateUUID()
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

func rulesToJson(groupRefId string, groupName string, groupType string, rules []interface{}) (result []RuleJSON, err error) {
	result = make([]RuleJSON, len(rules))

	for ruleIdx, r := range rules {
		r := r.(map[string]interface{})

		rj := &result[ruleIdx]

		rj.Type = groupType
		if groupType == "categorize" {
			rj.Ref_id = groupRefId
			rj.Name = groupName
		} else if groupType == "filter" {
			rj.To = groupRefId
		} else {
			return nil, fmt.Errorf("Unrecognized group type %s", groupType)
		}

		rj.Asset = stringOrNil(r["asset"])
		rj.Field = convertStringArray(r["field"])
		rj.Tag_field = convertStringArray(r["tag_field"])

		if r["condition"] != nil {
			rj.Condition = conditionsToJson(r["condition"].([]interface{}), stringOrNil(r["combine_with"]))
		} else {
			rj.Condition = nil
		}
	}
	return result, nil
}

func conditionsToJson(conditions []interface{}, combineWith string) (result *ConditionJSON) {
	if len(conditions) == 0 {
		return nil
	}
	result = new(ConditionJSON)
	result.Clauses = make([]ClauseJSON, len(conditions))
	result.Combine_with = combineWith
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

func constantToJson(tfConstant map[string]interface{}) (constantType string, constantItem ConstantItem) {
	constantType = tfConstant["constant_type"].(string)
	constantItem = ConstantItem{
		Ref_id: stringOrNil(tfConstant["ref_id"]),
		Name:   stringOrNil(tfConstant["name"]),
		Val:    stringOrNil(tfConstant["val"]),
	}
	if constantType == DynamicGroupType {
		blk_id := stringOrNil(tfConstant["blk_id"])
		constantItem.Blk_id = &blk_id
	}
	if stringOrNil(tfConstant["is_other"]) == "true" {
		constantItem.Is_other = "true"
	}
	return constantType, constantItem
}

func addOtherConstants(tfConstants []interface{}, constantsByType map[string]*ConstantJSON) error {
	// Add "other" constants
	// These are constants that have literally is_other == "true" or dynamic
	// groups with empty blk_ids
	for _, tfConstant := range tfConstants {
		tfConstant := tfConstant.(map[string]interface{})

		if tfConstant["is_other"].(string) == "true" ||
			(tfConstant["constant_type"].(string) == DynamicGroupType && tfConstant["blk_id"] == "") {

			constantType, constantItem := constantToJson(tfConstant)

			constant := constantsByType[constantType]
			if constant == nil {
				return fmt.Errorf("Unknown constant type %s", constantType)
			}
			constant.List = append(constant.List, constantItem)
		}
	}
	return nil
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
