package ccloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/v2/billing/masterdata/projects"
)

func billingProjectFlattenCostObject(co projects.CostObject) []map[string]interface{} {
	return []map[string]interface{}{{
		"inherited": co.Inherited,
		"name":      co.Name,
		"type":      co.Type,
	}}
}

func billingProjectExpandCostObject(raw interface{}) projects.CostObject {
	var co projects.CostObject

	if raw != nil {
		if v, ok := raw.([]interface{}); ok {
			for _, v := range v {
				if v, ok := v.(map[string]interface{}); ok {
					if v, ok := v["inherited"]; ok {
						co.Inherited = v.(bool)
					}
					if !co.Inherited {
						if v, ok := v["name"]; ok {
							co.Name = v.(string)
						}
						if v, ok := v["type"]; ok {
							co.Type = v.(string)
						}
					}

					return co
				}
			}
		}
	}

	return co
}

// replaceEmptyString is a helper function to replace empty string fields with
// another field.
func replaceEmptyString(d *schema.ResourceData, field string, b string) string {
	var v interface{}
	var ok bool
	if v, ok = getOkExists(d, field); !ok {
		return b
	}
	return v.(string)
}

// replaceEmptyBool is a helper function to replace empty string fields with
// another field.
func replaceEmptyBool(d *schema.ResourceData, field string, b bool) bool {
	var v interface{}
	var ok bool
	if v, ok = getOkExists(d, field); !ok {
		return b
	}
	return v.(bool)
}

func billingProjectExpandExtCertificationV1(raw interface{}) *projects.ExtCertification {
	v, ok := raw.([]interface{})
	if !ok {
		return nil
	}

	for _, v := range v {
		v, ok := v.(map[string]interface{})
		if !ok {
			return nil
		}
		extCertification := &projects.ExtCertification{}
		if v, ok := v["c5"].(bool); ok {
			extCertification.C5 = v
		}
		if v, ok := v["iso"].(bool); ok {
			extCertification.ISO = v
		}
		if v, ok := v["pci"].(bool); ok {
			extCertification.PCI = v
		}
		if v, ok := v["soc1"].(bool); ok {
			extCertification.SOC1 = v
		}
		if v, ok := v["soc2"].(bool); ok {
			extCertification.SOC2 = v
		}
		if v, ok := v["SOX"].(bool); ok {
			extCertification.SOX = v
		}
		//nolint:staticcheck // we need the first element
		return extCertification
	}

	return nil
}

func billingProjectFlattenExtCertificationV1(extCertification *projects.ExtCertification) []map[string]interface{} {
	if extCertification == nil {
		return nil
	}

	return []map[string]interface{}{{
		"c5":   extCertification.C5,
		"iso":  extCertification.ISO,
		"pci":  extCertification.PCI,
		"soc1": extCertification.SOC1,
		"soc2": extCertification.SOC2,
		"sox":  extCertification.SOX,
	}}
}
