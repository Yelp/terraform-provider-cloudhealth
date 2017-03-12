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

		for _, dg := range tfGroup["dynamic_group"].([]interface{}) {
			dg := dg.(map[string]interface{})
			blk_id := tfGroup["ref_id"].(string)
			constantItem := ConstantItem{
				Name:   dg["name"].(string),
				Ref_id: dg["ref_id"].(string),
				Blk_id: &blk_id,
				Val:    dg["val"].(string),
			}
			constantsByType[DynamicGroupType].List = append(constantsByType[DynamicGroupType].List, constantItem)
		}

		var constant *ConstantJSON

		for _, r := range tfGroup["rule"].([]interface{}) {
			r := r.(map[string]interface{})

			var rj RuleJSON

			rj.Type = r["type"].(string)
			if rj.Type == "categorize" {
				if constant != nil && constant != constantsByType[DynamicGroupBlockType] {
					return nil, fmt.Errorf("Cannot support mixed categorize and filter rules in one group: %s", tfGroup["name"])
				}
				constant = constantsByType[DynamicGroupBlockType]
				rj.Ref_id = tfGroup["ref_id"].(string)
			} else if rj.Type == "filter" {
				if constant != nil && constant != constantsByType[StaticGroupType] {
					return nil, fmt.Errorf("Cannot support mixed categorize and filter rules in one group: %s", tfGroup["name"])
				}
				constant = constantsByType[StaticGroupType]
				rj.To = tfGroup["ref_id"].(string)
			} else if rj.Type == "" {
				return nil, fmt.Errorf("rule type not set!")
			} else {
				return nil, fmt.Errorf("Unrecognized rule type %s", rj.Type)
			}

			rj.Asset = stringOrNil(r["asset"])
			rj.Field = convertStringArray(r["field"])
			rj.Tag_field = convertStringArray(r["tag_field"])

			if r["condition"] != nil && len(r["condition"].([]interface{})) > 0 {
				conditions := r["condition"].([]interface{})
				rj.Condition = new(ConditionJSON)
				rj.Condition.Clauses = make([]ClauseJSON, len(conditions))
				for idx, condition := range conditions {
					condition := condition.(map[string]interface{})
					rj.Condition.Clauses[idx] = ClauseJSON{
						Field:     convertStringArray(condition["field"]),
						Tag_field: convertStringArray(condition["tag_field"]),
						Op:        stringOrNil(condition["op"]),
						Val:       stringOrNil(condition["val"]),
					}
				}
			} else {
				rj.Condition = nil
			}

			pj.Schema.Rules = append(pj.Schema.Rules, rj)
		}

		constantItem := ConstantItem{
			Name:   tfGroup["name"].(string),
			Ref_id: tfGroup["ref_id"].(string),
		}
		if constant == nil {
			return nil, fmt.Errorf("Unknown group type %s", tfGroup["name"])
		}
		constant.List = append(constant.List, constantItem)
	}

	for _, otherGroup := range otherGroups {
		otherGroup := otherGroup.(map[string]interface{})
		constantType := otherGroup["constant_type"].(string)

		constantItem := ConstantItem{
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
		constant := constantsByType[constantType]
		if constant == nil {
			return nil, fmt.Errorf("Unknown constant type %s", constantType)
		}
		constant.List = append(constant.List, constantItem)
	}

	for _, constantGroup := range constantsByType {
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
