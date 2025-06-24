package sci

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceSCIArcAgentV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIArcAgentV1Read,

		// Terraform timeouts don't work in data sources.
		// However "Timeouts" has to be specified, otherwise "timeouts" argument below won't work.
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(0),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"agent_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"filter"},
				ValidateFunc:  validation.NoZeroValues,
			},

			"filter": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"agent_id"},
				ValidateFunc:  validation.NoZeroValues,
			},

			// Terraform timeouts don't work in data sources.
			// This is a workaround.
			"timeouts": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"read": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateTimeout,
						},
					},
				},
			},

			// computed attributes
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"project": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_with": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_by": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"all_tags": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"facts": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"facts_agents": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceSCIArcAgentV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	agentID := d.Get("agent_id").(string)
	filter := d.Get("filter").(string)

	timeout, err := arcAgentV1ParseTimeout(d.Get("timeouts"))
	if err != nil {
		return diag.Errorf("Error parsing the read timeout for sci_arc_job_v1: %s", err)
	}

	agent, err := arcSCIArcAgentV1WaitForAgent(ctx, arcClient, agentID, filter, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(agent.AgentID)

	arcSCIArcAgentV1ReadAgent(ctx, d, arcClient, agent, GetRegion(d, config))

	return nil
}
