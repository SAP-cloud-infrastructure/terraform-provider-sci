package sci

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/v2/automation/v1/runs"
)

func resourceSCIAutomationRunV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIAutomationRunV1Create,
		ReadContext:   resourceSCIAutomationRunV1Read,
		DeleteContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics { return nil },

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			"automation_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"selector": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			// Computed
			"automation_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"repository": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "This field is not returned by the Lyra API",
			},

			"repository_revision": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"automation_attributes": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"log": {
				Type:     schema.TypeString,
				Computed: true,
				// Don't print the huge log during the terraform plan/apply
				Sensitive: true,
			},

			"jobs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"project_id": {
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

			"owner": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceSCIAutomationRunV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	createOpts := runs.CreateOpts{
		AutomationID: d.Get("automation_id").(string),
		Selector:     d.Get("selector").(string),
	}

	log.Printf("[DEBUG] sci_automation_run_v1 create options: %#v", createOpts)

	run, err := runs.Create(ctx, automationClient, createOpts).Extract()
	if err != nil {
		return diag.Errorf("Error creating sci_automation_run_v1: %s", err)
	}

	d.SetId(run.ID)

	timeout := d.Timeout(schema.TimeoutCreate)
	target := []string{
		"completed",
		"failed",
	}
	pending := []string{
		"preparing",
		"executing",
	}
	err = waitForAutomationRunV1(ctx, automationClient, run.ID, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSCIAutomationRunV1Read(ctx, d, meta)
}

func resourceSCIAutomationRunV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Automation client: %s", err)
	}

	run, err := runs.Get(ctx, automationClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Unable to retrieve sci_automation_run_v1"))
	}

	_ = d.Set("automation_id", run.AutomationID)
	_ = d.Set("automation_name", run.AutomationName)
	_ = d.Set("selector", run.Selector)
	_ = d.Set("repository_revision", run.RepositoryRevision)

	automationAttributes, err := json.Marshal(run.AutomationAttributes)
	if err != nil {
		log.Printf("[DEBUG] resourceSCIAutomationRunV1Read: Cannot marshal run.AutomationAttributes: %s", err)
	}
	_ = d.Set("automation_attributes", string(automationAttributes))

	_ = d.Set("state", run.State)
	_ = d.Set("created_at", run.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", run.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("log", run.Log)
	_ = d.Set("jobs", run.Jobs)
	_ = d.Set("owner", flattenAutomationiOwnerV1(run.Owner))
	_ = d.Set("project_id", run.ProjectID)

	_ = d.Set("region", GetRegion(d, config))

	return nil
}

func flattenAutomationiOwnerV1(owner runs.Owner) []any {
	return []any{map[string]any{
		"id":          owner.ID,
		"name":        owner.Name,
		"domain_id":   owner.DomainID,
		"domain_name": owner.DomainName,
	}}
}

func waitForAutomationRunV1(ctx context.Context, automationClient *gophercloud.ServiceClient, id string, target []string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for %s run to become %v.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     target,
		Pending:    pending,
		Refresh:    automationRunV1GetState(ctx, automationClient, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func automationRunV1GetState(ctx context.Context, automationClient *gophercloud.ServiceClient, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		run, err := runs.Get(ctx, automationClient, id).Extract()
		if err != nil {
			return nil, "", fmt.Errorf("enable to retrieve %s sci_automation_run_v1: %v", id, err)
		}

		return run, run.State, nil
	}
}
