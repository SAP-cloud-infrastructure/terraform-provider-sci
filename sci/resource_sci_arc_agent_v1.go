package sci

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/v2/arc/v1/agents"
)

func resourceSCIArcAgentV1() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSCIArcAgentV1Read,
		CreateContext: resourceSCIArcAgentV1Create,
		UpdateContext: resourceSCIArcAgentV1Update,
		DeleteContext: resourceSCIArcAgentV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
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

			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
			},

			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			// Computed attributes
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

func resourceSCIArcAgentV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	agentID := d.Get("agent_id").(string)
	filter := d.Get("filter").(string)
	timeout := d.Timeout(schema.TimeoutCreate)

	agent, err := arcSCIArcAgentV1WaitForAgent(ctx, arcClient, agentID, filter, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(agent.AgentID)

	err = updateArcAgentTagsV1(ctx, arcClient, d.Id(), nil, d.Get("tags"))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSCIArcAgentV1Read(ctx, d, meta)
}

func resourceSCIArcAgentV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	agent, err := agents.Get(ctx, arcClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Unable to retrieve sci_arc_agent_v1"))
	}

	arcSCIArcAgentV1ReadAgent(ctx, d, arcClient, agent, GetRegion(d, config))

	return nil
}

func resourceSCIArcAgentV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	oldTags, newTags := d.GetChange("tags")
	err = updateArcAgentTagsV1(ctx, arcClient, d.Id(), oldTags, newTags)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSCIArcAgentV1Read(ctx, d, meta)
}

func resourceSCIArcAgentV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	if !d.Get("force_delete").(bool) {
		// Wait for the instance to delete before moving on.
		log.Printf("[DEBUG] Waiting for compute instance (%s) to delete", d.Id())

		computeClient, err := config.ComputeV2Client(ctx, GetRegion(d, config))
		if err != nil {
			return diag.Errorf("Error creating OpenStack compute client: %s", err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"ACTIVE", "SHUTOFF"},
			Target:     []string{"DELETED", "SOFT_DELETED"},
			Refresh:    serverV2StateRefreshFunc(ctx, computeClient, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutDelete),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("Error waiting for compute instance (%s) to delete: %v", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting sci_arc_agent_v1: %s", d.Id())
	err = agents.Delete(ctx, arcClient, d.Id()).ExtractErr()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Error deleting sci_arc_agent_v1"))
	}

	return nil
}
