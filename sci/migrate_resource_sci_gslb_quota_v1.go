package sci

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSCIGSLBQuotaV1V0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"datacenter": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"member": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"monitor": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"pool": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			// computed
			"in_use_datacenter": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_domain": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_member": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_monitor": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_pool": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceSCIGSLBQuotaV1StateUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	rawState["domain_akamai"] = rawState["domain"]
	rawState["in_use_domain_akamai"] = rawState["in_use_domain"]

	return rawState, nil
}
