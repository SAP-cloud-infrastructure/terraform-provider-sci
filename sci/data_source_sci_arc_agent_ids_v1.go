package sci

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gophercloud/utils/v2/terraform/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/v2/arc/v1/agents"
)

func dataSourceSCIArcAgentIDsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIArcAgentIDsV1Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"filter": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// computed attributes
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSCIArcAgentIDsV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	filter := d.Get("filter").(string)

	listOpts := agents.ListOpts{Filter: filter}

	log.Printf("[DEBUG] sci_arc_agent_ids_v1 list options: %#v", listOpts)

	allPages, err := agents.List(arcClient, listOpts).AllPages(ctx)
	if err != nil {
		return diag.Errorf("Unable to list sci_arc_agent_ids_v1: %s", err)
	}

	allAgents, err := agents.ExtractAgents(allPages)
	if err != nil {
		return diag.Errorf("Unable to retrieve sci_arc_agent_ids_v1: %s", err)
	}

	agentIDs := make([]string, 0, len(allAgents))
	for _, a := range allAgents {
		agentIDs = append(agentIDs, a.AgentID)
	}

	log.Printf("[DEBUG] Retrieved %d agents in sci_arc_agent_ids_v1: %+v", len(allAgents), allAgents)

	d.SetId(fmt.Sprintf("%d", hashcode.String(strings.Join(agentIDs, ""))))
	_ = d.Set("ids", agentIDs)

	return nil
}
