package cloudhealth

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ugorji/go/codec"
	"strconv"
)

func jsonToTF(rawData []byte, d *schema.ResourceData) error {
	// parse the json
	var pj PerspectiveJSON

	var jsonHandle codec.JsonHandle
	jsonHandle.ErrorIfNoField = true

	var dec *codec.Decoder = codec.NewDecoderBytes(rawData, &jsonHandle)
	err := dec.Decode(&pj)
	if err != nil {
		return fmt.Errorf("Unable to parse json for perspective %s because %s", d.Id(), err)
	}

	// Now load the json into TF schema

	d.Set("name", pj.Schema.Name)

	if v, err := strconv.ParseBool(pj.Schema.Include_in_reports); err == nil {
		err = d.Set("include_in_reports", v)
	}
	if err != nil {
		return err
	}

	groupByRef := jsonToGroups(pj)
	groups, err := populateRules(pj, groupByRef)
	if err != nil {
		return err
	}

	constants := buildConstants(pj)

	d.Set("group", groups)

	err = d.Set("constant", constants)
	if err != nil {
		return err
	}
	return nil
}

func jsonToGroups(pj PerspectiveJSON) (groupByRef map[string]Group) {
	groupByRef = make(map[string]Group)

	for _, constant := range pj.Schema.Constants {
		if constant.Type != StaticGroupType && constant.Type != DynamicGroupBlockType {
			continue
		}
		for _, constantGroup := range constant.List {
			if constantGroup.Is_other == "true" {
				// An "other" group, solely handled by buildConstants()
				continue
			}
			group := make(Group)
			group["name"] = constantGroup.Name
			group["ref_id"] = constantGroup.Ref_id
			group["rule"] = make([]map[string]interface{}, 0)
			if constant.Type == DynamicGroupBlockType {
				group["type"] = "categorize"
			} else {
				group["type"] = "filter"
			}
			groupByRef[constantGroup.Ref_id] = group
		}
	}
	return groupByRef
}

func populateRules(pj PerspectiveJSON, groupByRef map[string]Group) (groups []Group, err error) {
	groupByRefSeen := make(map[string]bool)
	groups = make([]Group, 0)
	for _, jsonRule := range pj.Schema.Rules {
		groupRef := jsonRule.To
		if groupRef == "" {
			groupRef = jsonRule.Ref_id
			if groupRef == "" {
				return nil, fmt.Errorf("Unable to find 'to' for rule for asset %s", jsonRule.Type)
			}
		}

		rule := make(map[string]interface{})
		group := groupByRef[groupRef]
		if group == nil {
			return nil, fmt.Errorf("Group reference %s not found", groupRef)
		}

		// Order the groups by order that the rules are seen.  CHT technically
		// allows the groups for rules to be interleaved, but this is horribly
		// confusing and not in the UI
		if groupByRefSeen[groupRef] == false {
			groups = append(groups, group)
			groupByRefSeen[groupRef] = true
		}

		group["rule"] = append(group["rule"].([]map[string]interface{}), rule)

		if jsonRule.Type != group["type"] {
			return nil, fmt.Errorf("Unknown rule type %s; expected %s", jsonRule.Type, group["type"])
		}
		rule["asset"] = jsonRule.Asset
		if jsonRule.Tag_field != nil {
			rule["tag_field"] = jsonRule.Tag_field
		}
		if jsonRule.Field != nil {
			rule["field"] = jsonRule.Field
		}

		if jsonRule.Condition != nil {
			rule["combine_with"] = jsonRule.Condition.Combine_with
			jsonClauses := jsonRule.Condition.Clauses
			if jsonClauses != nil {
				rule["condition"] = buildCondition(jsonClauses)
			}
		}
	}

	return groups, nil
}

func buildCondition(jsonClauses []ClauseJSON) (clauses []map[string]interface{}) {
	clauses = make([]map[string]interface{}, len(jsonClauses))

	for idx, jsonClause := range jsonClauses {
		clause := make(map[string]interface{})
		clauses[idx] = clause

		if jsonClause.Tag_field != nil {
			clause["tag_field"] = jsonClause.Tag_field
		}
		if jsonClause.Field != nil {
			clause["field"] = jsonClause.Field
		}
		clause["op"] = jsonClause.Op
		clause["val"] = jsonClause.Val
	}
	return clauses
}

func buildConstants(pj PerspectiveJSON) []Group {
	result := make([]Group, 0)
	for _, jsonConstant := range pj.Schema.Constants {
		for _, jsonConstantGroup := range jsonConstant.List {
			constant := Group{
				"constant_type": jsonConstant.Type,
				"ref_id":        jsonConstantGroup.Ref_id,
				"name":          jsonConstantGroup.Name,
				"val":           jsonConstantGroup.Val,
				"is_other":      jsonConstantGroup.Is_other,
			}
			if jsonConstantGroup.Blk_id != nil {
				constant["blk_id"] = *jsonConstantGroup.Blk_id
			}

			result = append(result, constant)
		}
	}

	return result
}
