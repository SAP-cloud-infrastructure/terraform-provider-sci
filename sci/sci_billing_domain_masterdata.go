package sci

import (
	"github.com/sapcc/gophercloud-sapcc/v2/billing/masterdata/domains"
)

func billingDomainFlattenCostObject(co domains.CostObject) []map[string]any {
	return []map[string]any{{
		"projects_can_inherit": co.ProjectsCanInherit,
		"name":                 co.Name,
		"type":                 co.Type,
	}}
}

func billingDomainExpandCostObject(raw any) domains.CostObject {
	var co domains.CostObject

	if raw != nil {
		if v, ok := raw.([]any); ok {
			for _, v := range v {
				if v, ok := v.(map[string]any); ok {
					if v, ok := v["projects_can_inherit"]; ok {
						co.ProjectsCanInherit = v.(bool)
					}
					if v, ok := v["name"]; ok {
						co.Name = v.(string)
					}
					if v, ok := v["type"]; ok {
						co.Type = v.(string)
					}

					return co
				}
			}
		}
	}

	return co
}
