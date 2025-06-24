package sci

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/sapcc/kubernikus/pkg/api/models"
)

// CheckDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeleted(d *schema.ResourceData, err error, msg string) error {
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		d.SetId("")
		return nil
	}

	return fmt.Errorf("%s %s: %v", msg, d.Id(), err)
}

// GetRegion returns the region that was specified in the resource. If a
// region was not set, the provider-level region is checked. The provider-level
// region can either be set by the region argument or by OS_REGION_NAME.
func GetRegion(d *schema.ResourceData, config *Config) string {
	if v, ok := d.GetOk("region"); ok {
		return v.(string)
	}

	return config.Region
}

// sliceContains returns true if the element exists in the slice.
func sliceContains[T comparable](sl []T, el T) bool {
	for _, s := range sl {
		if s == el {
			return true
		}
	}
	return false
}

// strSliceContains returns true if the string exists in given slice, ignore case.
func strSliceContains(sl []string, str string) bool {
	str = strings.ToLower(str)
	for _, s := range sl {
		if strings.ToLower(s) == str {
			return true
		}
	}
	return false
}

func expandToMapStringString(v map[string]any) map[string]string {
	m := make(map[string]string)
	for key, val := range v {
		if strVal, ok := val.(string); ok {
			m[key] = strVal
		}
		if strVal, ok := val.(bool); ok {
			m[key] = fmt.Sprintf("%t", strVal)
		}
	}

	return m
}

func expandToStringSlice(v []any) []string {
	s := make([]string, len(v))
	for i, val := range v {
		if strVal, ok := val.(string); ok {
			s[i] = strVal
		}
	}

	return s
}

func expandToStrFmtUUIDSlice(v []any) []strfmt.UUID {
	s := make([]strfmt.UUID, len(v))
	for i, val := range v {
		if strVal, ok := val.(string); ok {
			s[i] = strfmt.UUID(strVal)
		}
	}

	return s
}

func expandToNodePoolConfig(v []any) *models.NodePoolConfig {
	c := new(models.NodePoolConfig)
	for _, val := range v {
		if mapVal, ok := val.(map[string]any); ok {
			if v, ok := mapVal["allow_reboot"].(bool); ok {
				c.AllowReboot = &v
			}
			if v, ok := mapVal["allow_replace"].(bool); ok {
				c.AllowReplace = &v
			}
		}
	}

	return c
}

// from https://github.com/terraform-provider-openstack/terraform-provider-openstack/blob/74d82f6ce503df74a5e63ac2491e837dc296a82b/openstack/util.go#L153
func expandObjectTags(d *schema.ResourceData) []string {
	rawTags := d.Get("tags").(*schema.Set).List()
	tags := make([]string, len(rawTags))

	for i, raw := range rawTags {
		tags[i] = raw.(string)
	}

	return tags
}

func normalizeJSONString(v any) string {
	json, _ := structure.NormalizeJsonString(v)
	return json
}

func validateURL(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := url.ParseRequestURI(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q URL is not valid: %v", k, err))
	}
	return
}

func validateJSONObject(v any, k string) ([]string, []error) {
	if v == nil || v.(string) == "" {
		return nil, []error{fmt.Errorf("%q value must not be empty", k)}
	}

	var j map[string]any
	s := v.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be a JSON object: %v", k, err)}
	}

	return nil, nil
}

func validateJSONArray(v any, k string) ([]string, []error) {
	if v == nil || v.(string) == "" {
		return nil, []error{fmt.Errorf("%q value must not be empty", k)}
	}

	var j []any
	s := v.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be a JSON array: %v", k, err)}
	}

	return nil, nil
}

func validateTimeout(v any, k string) ([]string, []error) {
	if v == nil || v.(string) == "" {
		return nil, []error{fmt.Errorf("%q value must not be empty", k)}
	}

	_, err := time.ParseDuration(v.(string))
	if err != nil {
		return nil, []error{fmt.Errorf("%q: %v", k, err)}
	}

	return nil, nil
}

func diffSuppressJSONObject(k, old, new string, d *schema.ResourceData) bool {
	if strSliceContains([]string{"{}", "null", ""}, old) &&
		strSliceContains([]string{"{}", "null", ""}, new) {
		return true
	}
	return false
}

func diffSuppressJSONArray(k, old, new string, d *schema.ResourceData) bool {
	if strSliceContains([]string{"[]", "null", ""}, old) &&
		strSliceContains([]string{"[]", "null", ""}, new) {
		return true
	}
	return false
}

func validateKubernetesVersion(v any, k string) ([]string, []error) {
	if err := validate.Pattern("version", "", v.(string), `^[0-9]+\.[0-9]+\.[0-9]+$`); err != nil {
		return nil, []error{fmt.Errorf("invalid version (%s) specified for Kubernikus cluster: %v", v.(string), err)}
	}

	return nil, nil
}

func removePrefixIPAddress(ip string) string {
	res, _, _ := net.ParseCIDR(ip)
	if res == nil {
		ip = ip + "/32"
		res, _, _ = net.ParseCIDR(ip)
		if res == nil {
			return ""
		}
	}
	return res.String()
}

func expandToStrFmtIPv4Slice(v []any) []strfmt.IPv4 {
	s := make([]strfmt.IPv4, len(v))
	for i, val := range v {
		if strVal, ok := val.(string); ok {
			s[i] = strfmt.IPv4(removePrefixIPAddress(strVal))
		}
	}

	return s
}

func flattenToStrFmtIPv4Slice(v []strfmt.IPv4) []string {
	s := make([]string, len(v))
	for i, val := range v {
		s[i] = removePrefixIPAddress(string(val))
	}

	return s
}

func ptr[T any](v T) *T {
	return &v
}

func ptrValue[T any](p *T) T {
	if p != nil {
		return *p
	}
	var t T
	return t
}

// parsePairedIDs is a helper function that parses a raw ID into two
// separate IDs. This is useful for resources that have a parent/child
// relationship.
func parsePairedIDs(id string, res string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unable to determine %s ID from raw ID: %s", res, id)
	}

	return parts[0], parts[1], nil
}

// getOkExists is a helper function that replaces the deprecated GetOkExists
// schema method. It returns the value of the key if it exists in the
// configuration, along with a boolean indicating if the key exists.
func getOkExists(d *schema.ResourceData, key string) (any, bool) {
	v := d.GetRawConfig().GetAttr(key)
	if v.IsNull() {
		return nil, false
	}
	return d.Get(key), true
}
