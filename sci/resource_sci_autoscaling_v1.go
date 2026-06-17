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
	"go.xyrillian.de/gg/options"
)

func resourceSCIAutoscalingV1() *schema.Resource {
	thresholdSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metric": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(castellum.SingularUsageMetric),
			},
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
		ReadContext:   resourceSCIAutoscalingV1Read,
		CreateContext: resourceSCIAutoscalingV1Create,
		UpdateContext: resourceSCIAutoscalingV1Update,
		DeleteContext: resourceSCIAutoscalingV1Delete,
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
				Elem:     thresholdSchema,
			},

			"high_threshold": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     thresholdSchema,
			},

			"critical_threshold": {
				Type:     schema.TypeList,
				Optional: true,
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

func resourceSCIAutoscalingV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	opts := sciAutoscalingV1BuildOpts(d)

	log.Printf("[DEBUG] sci_autoscaling_v1 create options: %#v", opts)

	err = resources.Create(ctx, castellumClient, projectID, resourceType, opts).ExtractErr()
	if err != nil {
		return diag.Errorf("Error creating sci_autoscaling_v1 %s/%s: %s", projectID, resourceType, err)
	}

	d.SetId(sciAutoscalingV1BuildID(projectID, resourceType))

	return resourceSCIAutoscalingV1Read(ctx, d, meta)
}

func resourceSCIAutoscalingV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	castellumClient, err := config.castellumV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Castellum client: %s", err)
	}

	projectID, resourceType, err := sciAutoscalingV1ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resource, err := resources.Get(ctx, castellumClient, projectID, resourceType).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Unable to retrieve sci_autoscaling_v1"))
	}

	sciAutoscalingV1SetState(d, resource, projectID, resourceType, GetRegion(d, config))

	return nil
}

func resourceSCIAutoscalingV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	castellumClient, err := config.castellumV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Castellum client: %s", err)
	}

	projectID, resourceType, err := sciAutoscalingV1ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	opts := sciAutoscalingV1BuildOpts(d)

	log.Printf("[DEBUG] sci_autoscaling_v1 update options: %#v", opts)

	// The castellum API uses PUT for both create and update (idempotent upsert).
	err = resources.Create(ctx, castellumClient, projectID, resourceType, opts).ExtractErr()
	if err != nil {
		return diag.Errorf("Error updating sci_autoscaling_v1 %s/%s: %s", projectID, resourceType, err)
	}

	return resourceSCIAutoscalingV1Read(ctx, d, meta)
}

func resourceSCIAutoscalingV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	castellumClient, err := config.castellumV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Castellum client: %s", err)
	}

	projectID, resourceType, err := sciAutoscalingV1ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting sci_autoscaling_v1 %s/%s", projectID, resourceType)

	err = resources.Delete(ctx, castellumClient, projectID, resourceType).ExtractErr()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Error deleting sci_autoscaling_v1"))
	}

	return nil
}

// sciAutoscalingV1BuildOpts constructs CreateOpts from schema data.
func sciAutoscalingV1BuildOpts(d *schema.ResourceData) resources.CreateOpts {
	return resources.CreateOpts{
		LowThreshold:      sciAutoscalingV1ExpandThreshold(d.Get("low_threshold")),
		HighThreshold:     sciAutoscalingV1ExpandThreshold(d.Get("high_threshold")),
		CriticalThreshold: sciAutoscalingV1ExpandThreshold(d.Get("critical_threshold")),
		SizeConstraints:   sciAutoscalingV1ExpandSizeConstraints(d.Get("size_constraints")),
		SizeSteps:         sciAutoscalingV1ExpandSizeSteps(d.Get("size_steps")),
	}
}

func sciAutoscalingV1ExpandThreshold(raw any) option.Option[castellum.Threshold] {
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return option.None[castellum.Threshold]()
	}

	usageValues := castellum.UsageValues{}
	var delaySeconds uint32

	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		metric := castellum.UsageMetric(m["metric"].(string))
		usageValues[metric] = m["usage_percent"].(float64)
		delaySeconds = uint32(m["delay_seconds"].(int))
	}

	if len(usageValues) == 0 {
		return option.None[castellum.Threshold]()
	}

	return option.Some(castellum.Threshold{
		UsagePercent: usageValues,
		DelaySeconds: delaySeconds,
	})
}

func sciAutoscalingV1ExpandSizeConstraints(raw any) option.Option[castellum.SizeConstraints] {
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return option.None[castellum.SizeConstraints]()
	}
	m, ok := list[0].(map[string]any)
	if !ok {
		return option.None[castellum.SizeConstraints]()
	}
	sc := castellum.SizeConstraints{
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
	return option.Some(sc)
}

func sciAutoscalingV1ExpandSizeSteps(raw any) option.Option[castellum.SizeSteps] {
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return option.None[castellum.SizeSteps]()
	}
	m, ok := list[0].(map[string]any)
	if !ok {
		return option.None[castellum.SizeSteps]()
	}
	return option.Some(castellum.SizeSteps{
		Percent: m["percent"].(float64),
		Single:  m["single"].(bool),
	})
}

func sciAutoscalingV1FlattenThreshold(t castellum.Threshold) []map[string]any {
	result := make([]map[string]any, 0, len(t.UsagePercent))
	for metric, pct := range t.UsagePercent {
		result = append(result, map[string]any{
			"metric":        string(metric),
			"usage_percent": pct,
			"delay_seconds": int(t.DelaySeconds),
		})
	}
	return result
}

func sciAutoscalingV1FlattenSizeConstraints(sc castellum.SizeConstraints) []map[string]any {
	m := map[string]any{
		"minimum_free_is_critical": sc.MinimumFreeIsCritical,
		"minimum":                  int(sc.Minimum.UnwrapOr(0)),
		"minimum_free":             int(sc.MinimumFree.UnwrapOr(0)),
	}
	if v, ok := sc.Maximum.Unpack(); ok {
		m["maximum"] = int(v)
	}
	return []map[string]any{m}
}

func sciAutoscalingV1FlattenSizeSteps(ss castellum.SizeSteps) []map[string]any {
	return []map[string]any{
		{
			"percent": ss.Percent,
			"single":  ss.Single,
		},
	}
}

// sciAutoscalingV1BuildID formats the Terraform resource ID.
func sciAutoscalingV1BuildID(projectID, resourceType string) string {
	return projectID + "/" + resourceType
}

// sciAutoscalingV1ParseID splits the Terraform resource ID.
func sciAutoscalingV1ParseID(id string) (projectID, resourceType string, err error) {
	projectID, resourceType, ok := strings.Cut(id, "/")
	if !ok || projectID == "" || resourceType == "" {
		return "", "", fmt.Errorf("invalid sci_autoscaling_v1 ID %q: expected <project_id>/<resource_type>", id)
	}
	return projectID, resourceType, nil
}

// sciAutoscalingV1SetState populates schema.ResourceData from a fetched Resource.
func sciAutoscalingV1SetState(d *schema.ResourceData, resource castellum.Resource, projectID, resourceType, region string) {
	_ = d.Set("region", region)
	_ = d.Set("project_id", projectID)
	_ = d.Set("resource_type", resourceType)

	_ = d.Set("low_threshold", options.Map(resource.LowThreshold, sciAutoscalingV1FlattenThreshold).UnwrapOr([]map[string]any{}))
	_ = d.Set("high_threshold", options.Map(resource.HighThreshold, sciAutoscalingV1FlattenThreshold).UnwrapOr([]map[string]any{}))
	_ = d.Set("critical_threshold", options.Map(resource.CriticalThreshold, sciAutoscalingV1FlattenThreshold).UnwrapOr([]map[string]any{}))
	_ = d.Set("size_constraints", options.Map(resource.SizeConstraints, sciAutoscalingV1FlattenSizeConstraints).UnwrapOr([]map[string]any{}))

	if resource.SizeSteps.Percent != 0 || resource.SizeSteps.Single {
		_ = d.Set("size_steps", sciAutoscalingV1FlattenSizeSteps(resource.SizeSteps))
	} else {
		_ = d.Set("size_steps", []map[string]any{})
	}
}
