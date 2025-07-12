package sci

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-provider-openstack/utils/v2/hashcode"
)

func dataSourceSCIArcJobIDsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIArcJobIDsV1Read,

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

func dataSourceSCIArcJobIDsV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	jobs, err := arcSCIArcJobV1Filter(ctx, d, arcClient, "sci_arc_job_ids_v1")
	if err != nil {
		return diag.FromErr(err)
	}

	jobIDs := make([]string, 0, len(jobs))
	for _, j := range jobs {
		jobIDs = append(jobIDs, j.RequestID)
	}

	log.Printf("[DEBUG] Retrieved %d jobs in sci_arc_job_ids_v1: %+v", len(jobs), jobs)

	d.SetId(fmt.Sprintf("%d", hashcode.String(strings.Join(jobIDs, ""))))
	_ = d.Set("ids", jobIDs)

	return nil
}
