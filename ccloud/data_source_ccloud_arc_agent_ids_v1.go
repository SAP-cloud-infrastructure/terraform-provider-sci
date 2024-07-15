package ccloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/v2/arc/v1/agents"

	"github.com/gophercloud/utils/v2/terraform/hashcode"
)

func dataSourceCCloudArcAgentIDsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCCloudArcAgentIDsV1Read,

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

func dataSourceCCloudArcAgentIDsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	filter := d.Get("filter").(string)

	listOpts := agents.ListOpts{Filter: filter}

	log.Printf("[DEBUG] ccloud_arc_agent_ids_v1 list options: %#v", listOpts)

	allPages, err := agents.List(arcClient, listOpts).AllPages(ctx)
	if err != nil {
		return diag.Errorf("Unable to list ccloud_arc_agent_ids_v1: %s", err)
	}

	allAgents, err := agents.ExtractAgents(allPages)
	if err != nil {
		return diag.Errorf("Unable to retrieve ccloud_arc_agent_ids_v1: %s", err)
	}

	agentIDs := make([]string, 0, len(allAgents))
	for _, a := range allAgents {
		agentIDs = append(agentIDs, a.AgentID)
	}

	log.Printf("[DEBUG] Retrieved %d agents in ccloud_arc_agent_ids_v1: %+v", len(allAgents), allAgents)

	d.SetId(fmt.Sprintf("%d", hashcode.String(strings.Join(agentIDs, ""))))
	d.Set("ids", agentIDs)

	return nil
}
