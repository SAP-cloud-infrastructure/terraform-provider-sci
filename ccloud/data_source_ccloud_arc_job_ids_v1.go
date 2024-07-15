package ccloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/gophercloud/utils/v2/terraform/hashcode"
)

func dataSourceCCloudArcJobIDsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCCloudArcJobIDsV1Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"agent_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 86400),
			},

			"agent": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"chef", "execute",
				}, false),
			},

			"action": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"script", "zero", "tarball",
				}, false),
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"queued", "executing", "failed", "complete",
				}, false),
			},

			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCCloudArcJobIDsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	jobs, err := arcCCloudArcJobV1Filter(ctx, d, arcClient, "ccloud_arc_job_ids_v1")
	if err != nil {
		return diag.FromErr(err)
	}

	jobIDs := make([]string, 0, len(jobs))
	for _, j := range jobs {
		jobIDs = append(jobIDs, j.RequestID)
	}

	log.Printf("[DEBUG] Retrieved %d jobs in ccloud_arc_job_ids_v1: %+v", len(jobs), jobs)

	d.SetId(fmt.Sprintf("%d", hashcode.String(strings.Join(jobIDs, ""))))
	d.Set("ids", jobIDs)

	return nil
}
