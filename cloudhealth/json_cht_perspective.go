package cloudhealth

type ClauseJSON struct {
	Field     []string `json:"field,omitempty"`
	Tag_field []string `json:"tag_field,omitempty"`
	Op        string   `json:"op,omitempty"`
	Val       string   `json:"val,omitempty"`
}

type ConditionJSON struct {
	Combine_with string       `json:"combine_with,omitempty"`
	Clauses      []ClauseJSON `json:"clauses,omitempty"`
}

type RuleJSON struct {
	Type      string         `json:"type,omitempty"`
	Asset     string         `json:"asset,omitempty"`
	To        string         `json:"to,omitempty"`
	Ref_id    string         `json:"ref_id,omitempty"`    // for type='categorize'
	Field     []string       `json:"field,omitempty"`     // for type='categorize'
	Tag_field []string       `json:"tag_field,omitempty"` // for type='categorize'
	Condition *ConditionJSON `json:"condition,omitempty"`
}

type ConstantItem struct {
	Ref_id   string  `json:"ref_id,omitempty"`
	Blk_id   *string `json:"blk_id,omitempty"` // for Dynamic Groups
	Name     string  `json:"name,omitempty"`
	Val      string  `json:"val,omitempty"`      // for Dynamic Groups
	Is_other string  `json:"is_other,omitempty"` // the "Other" for Static Groups
}

type ConstantJSON struct {
	Type string         `json:"type,omitempty"`
	List []ConstantItem `json:"list,omitempty"`
}

type PerspectiveJSON struct {
	Schema struct {
		Name               string         `json:"name"`
		Include_in_reports string         `json:"include_in_reports"`
		Rules              []RuleJSON     `json:"rules"`
		Constants          []ConstantJSON `json:"constants"`
		Merges             []interface{}  `json:"merges"` // Not supported
	} `json:"schema"`
}

type Group map[string]interface{}

const StaticGroupType = "Static Group"
const DynamicGroupType = "Dynamic Group"
const DynamicGroupBlockType = "Dynamic Group Block"

func NewConstantJSON(t string) (constant *ConstantJSON) {
	constant = new(ConstantJSON)
	constant.Type = t
	constant.List = make([]ConstantItem, 0)
	return constant
}
