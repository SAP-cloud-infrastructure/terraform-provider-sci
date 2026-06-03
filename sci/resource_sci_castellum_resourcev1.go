// SPDX-FileCopyrightText: 2020-2026 SAP SE or an SAP affiliate company
// SPDX-FileCopyrightText: 2026 Dexter Le <dextersydney2001@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package sci

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/go-api-declarations/castellum"
	"github.com/sapcc/gophercloud-sapcc/v2/castellum/v1/resources"
	"go.xyrillian.de/gg/option"
)

func resourceSCICastellumResourceV1() *schema.Resource {
	thresholdSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"usage_percent": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"delay_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
		},
	}

	return &schema.Resource{
		ReadContext:   resourceSCICastellumResourceV1Read,
		CreateContext: resourceSCICastellumResourceV1Create,
		UpdateContext: resourceSCICastellumResourceV1Update,
		DeleteContext: resourceSCICastellumResourceV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"resource_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"low_threshold": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     thresholdSchema,
			},

			"high_threshold": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     thresholdSchema,
			},

			"critical_threshold": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     thresholdSchema,
			},

			"size_constraints": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minimum": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"maximum": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"minimum_free": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"minimum_free_is_critical": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"size_steps": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"percent": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"single": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

func resourceSCICastellumResourceV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	castellumClient, err := config.castellumV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Castellum client: %s", err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		identityClient, err := config.IdentityV3Client(ctx, GetRegion(d, config))
		if err != nil {
			return diag.Errorf("Error creating OpenStack identity client: %s", err)
		}
		tokenDetails, err := getTokenDetails(ctx, identityClient)
		if err != nil {
			return diag.FromErr(err)
		}
		if tokenDetails.project == nil {
			return diag.Errorf("Error resolving project_id from token scope: no project in token")
		}
		projectID = tokenDetails.project.ID
	}

	resourceType := d.Get("resource_type").(string)

	opts := castellumSCICastellumResourceV1BuildOpts(d)

	log.Printf("[DEBUG] sci_castellum_resource_v1 create options: %#v", opts)

	err = resources.Create(ctx, castellumClient, projectID, resourceType, opts).ExtractErr()
	if err != nil {
		return diag.Errorf("Error creating sci_castellum_resource_v1 %s/%s: %s", projectID, resourceType, err)
	}

	d.SetId(castellumSCICastellumResourceV1BuildID(projectID, resourceType))

	return resourceSCICastellumResourceV1Read(ctx, d, meta)
}

func resourceSCICastellumResourceV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	castellumClient, err := config.castellumV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Castellum client: %s", err)
	}

	projectID, resourceType, err := castellumSCICastellumResourceV1ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resource, err := resources.Get(ctx, castellumClient, projectID, resourceType).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Unable to retrieve sci_castellum_resource_v1"))
	}

	castellumSCICastellumResourceV1SetState(d, resource, projectID, resourceType, GetRegion(d, config))

	return nil
}

func resourceSCICastellumResourceV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	castellumClient, err := config.castellumV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Castellum client: %s", err)
	}

	projectID, resourceType, err := castellumSCICastellumResourceV1ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	opts := castellumSCICastellumResourceV1BuildOpts(d)

	log.Printf("[DEBUG] sci_castellum_resource_v1 update options: %#v", opts)

	// The castellum API uses PUT for both create and update (idempotent upsert).
	err = resources.Create(ctx, castellumClient, projectID, resourceType, opts).ExtractErr()
	if err != nil {
		return diag.Errorf("Error updating sci_castellum_resource_v1 %s/%s: %s", projectID, resourceType, err)
	}

	return resourceSCICastellumResourceV1Read(ctx, d, meta)
}

func resourceSCICastellumResourceV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	castellumClient, err := config.castellumV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Castellum client: %s", err)
	}

	projectID, resourceType, err := castellumSCICastellumResourceV1ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting sci_castellum_resource_v1 %s/%s", projectID, resourceType)

	err = resources.Delete(ctx, castellumClient, projectID, resourceType).ExtractErr()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Error deleting sci_castellum_resource_v1"))
	}

	return nil
}

// castellumSCICastellumResourceV1BuildOpts constructs CreateOpts from schema data.
func castellumSCICastellumResourceV1BuildOpts(d *schema.ResourceData) resources.CreateOpts {
	opts := resources.CreateOpts{}

	if v, ok := d.GetOk("low_threshold"); ok {
		if t := castellumSCICastellumResourceV1ExpandThreshold(v); t != nil {
			opts.LowThreshold = option.Some(*t)
		}
	}
	if v, ok := d.GetOk("high_threshold"); ok {
		if t := castellumSCICastellumResourceV1ExpandThreshold(v); t != nil {
			opts.HighThreshold = option.Some(*t)
		}
	}
	if v, ok := d.GetOk("critical_threshold"); ok {
		if t := castellumSCICastellumResourceV1ExpandThreshold(v); t != nil {
			opts.CriticalThreshold = option.Some(*t)
		}
	}
	if v, ok := d.GetOk("size_constraints"); ok {
		if sc := castellumSCICastellumResourceV1ExpandSizeConstraints(v); sc != nil {
			opts.SizeConstraints = option.Some(*sc)
		}
	}
	if v, ok := d.GetOk("size_steps"); ok {
		if ss := castellumSCICastellumResourceV1ExpandSizeSteps(v); ss != nil {
			opts.SizeSteps = option.Some(*ss)
		}
	}

	return opts
}

func castellumSCICastellumResourceV1ExpandThreshold(raw any) *castellum.Threshold {
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	m, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	return &castellum.Threshold{
		UsagePercent: castellum.UsageValues{
			castellum.SingularUsageMetric: m["usage_percent"].(float64),
		},
		DelaySeconds: uint32(m["delay_seconds"].(int)),
	}
}

func castellumSCICastellumResourceV1ExpandSizeConstraints(raw any) *castellum.SizeConstraints {
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	m, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	sc := &castellum.SizeConstraints{
		MinimumFreeIsCritical: m["minimum_free_is_critical"].(bool),
	}
	if v := m["minimum"].(int); v != 0 {
		sc.Minimum = option.Some(uint64(v))
	}
	if v := m["maximum"].(int); v != 0 {
		sc.Maximum = option.Some(uint64(v))
	}
	if v := m["minimum_free"].(int); v != 0 {
		sc.MinimumFree = option.Some(uint64(v))
	}
	return sc
}

func castellumSCICastellumResourceV1ExpandSizeSteps(raw any) *castellum.SizeSteps {
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	m, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}
	return &castellum.SizeSteps{
		Percent: m["percent"].(float64),
		Single:  m["single"].(bool),
	}
}

func castellumSCICastellumResourceV1FlattenThreshold(t castellum.Threshold) []map[string]any {
	return []map[string]any{
		{
			"usage_percent": t.UsagePercent[castellum.SingularUsageMetric],
			"delay_seconds": int(t.DelaySeconds),
		},
	}
}

func castellumSCICastellumResourceV1FlattenSizeConstraints(sc castellum.SizeConstraints) []map[string]any {
	m := map[string]any{
		"minimum_free_is_critical": sc.MinimumFreeIsCritical,
	}
	if v, ok := sc.Minimum.Unpack(); ok {
		m["minimum"] = int(v)
	} else {
		m["minimum"] = 0
	}
	if v, ok := sc.Maximum.Unpack(); ok {
		m["maximum"] = int(v)
	} else {
		m["maximum"] = 0
	}
	if v, ok := sc.MinimumFree.Unpack(); ok {
		m["minimum_free"] = int(v)
	} else {
		m["minimum_free"] = 0
	}
	return []map[string]any{m}
}

func castellumSCICastellumResourceV1FlattenSizeSteps(ss castellum.SizeSteps) []map[string]any {
	return []map[string]any{
		{
			"percent": ss.Percent,
			"single":  ss.Single,
		},
	}
}

// castellumSCICastellumResourceV1BuildID formats the Terraform resource ID.
func castellumSCICastellumResourceV1BuildID(projectID, resourceType string) string {
	return projectID + "/" + resourceType
}

// castellumSCICastellumResourceV1ParseID splits the Terraform resource ID.
func castellumSCICastellumResourceV1ParseID(id string) (projectID, resourceType string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid sci_castellum_resource_v1 ID %q: expected <project_id>/<resource_type>", id)
	}
	return parts[0], parts[1], nil
}

// castellumSCICastellumResourceV1SetState populates schema.ResourceData from a fetched Resource.
func castellumSCICastellumResourceV1SetState(d *schema.ResourceData, resource castellum.Resource, projectID, resourceType, region string) {
	_ = d.Set("region", region)
	_ = d.Set("project_id", projectID)
	_ = d.Set("resource_type", resourceType)

	if t, ok := resource.LowThreshold.Unpack(); ok {
		_ = d.Set("low_threshold", castellumSCICastellumResourceV1FlattenThreshold(t))
	} else {
		_ = d.Set("low_threshold", []map[string]any{})
	}

	if t, ok := resource.HighThreshold.Unpack(); ok {
		_ = d.Set("high_threshold", castellumSCICastellumResourceV1FlattenThreshold(t))
	} else {
		_ = d.Set("high_threshold", []map[string]any{})
	}

	if t, ok := resource.CriticalThreshold.Unpack(); ok {
		_ = d.Set("critical_threshold", castellumSCICastellumResourceV1FlattenThreshold(t))
	} else {
		_ = d.Set("critical_threshold", []map[string]any{})
	}

	if sc, ok := resource.SizeConstraints.Unpack(); ok {
		_ = d.Set("size_constraints", castellumSCICastellumResourceV1FlattenSizeConstraints(sc))
	} else {
		_ = d.Set("size_constraints", []map[string]any{})
	}

	if resource.SizeSteps.Percent != 0 || resource.SizeSteps.Single {
		_ = d.Set("size_steps", castellumSCICastellumResourceV1FlattenSizeSteps(resource.SizeSteps))
	} else {
		_ = d.Set("size_steps", []map[string]any{})
	}
}
