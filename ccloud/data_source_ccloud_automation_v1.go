package ccloud

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/v2/automation/v1/automations"
)

func dataSourceCCloudAutomationV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCCloudAutomationV1Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"repository": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"repository_revision": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Script", "Chef",
				}, false),
			},

			"tags": {
				Type:       schema.TypeMap,
				Computed:   true,
				Deprecated: "This field is not supported by the Lyra API",
			},

			// Chef
			"run_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"chef_attributes": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"log_level": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"debug": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"chef_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// Script
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"arguments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"environment": {
				Type:     schema.TypeMap,
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

			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"repository_authentication_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceCCloudAutomationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Automation client: %s", err)
	}

	allPages, err := automations.List(automationClient, automations.ListOpts{}).AllPages(ctx)
	if err != nil {
		return diag.Errorf("Unable to list ccloud_automation_v1: %s", err)
	}

	allAutomations, err := automations.ExtractAutomations(allPages)
	if err != nil {
		return diag.Errorf("Unable to retrieve ccloud_automation_v1: %s", err)
	}

	if len(allAutomations) == 0 {
		return diag.Errorf("No ccloud_automation_v1 found")
	}

	var automations []automations.Automation
	var v interface{}
	var debugExists, debug bool

	if v, debugExists = d.GetOkExists("debug"); debugExists {
		debug = v.(bool)
	}
	name := d.Get("name").(string)
	repository := d.Get("repository").(string)
	repositoryRevision := d.Get("repository_revision").(string)
	timeout := d.Get("timeout").(int)
	automationType := d.Get("type").(string)
	chefVersion := d.Get("chef_version").(string)
	path := d.Get("path").(string)

	for _, automation := range allAutomations {
		found := true
		if found && len(name) > 0 && automation.Name != name {
			found = false
		}
		if found && len(repository) > 0 && automation.Repository != repository {
			found = false
		}
		if found && len(repositoryRevision) > 0 && automation.RepositoryRevision != repositoryRevision {
			found = false
		}
		if found && timeout > 0 && automation.Timeout != timeout {
			found = false
		}
		if found && len(automationType) > 0 && automation.Type != automationType {
			found = false
		}
		if found && debugExists && automation.Debug != debug {
			found = false
		}
		if found && len(chefVersion) > 0 && automation.ChefVersion != chefVersion {
			found = false
		}
		if found && len(path) > 0 && automation.Path != path {
			found = false
		}

		if found {
			automations = append(automations, automation)
		}
	}

	if len(automations) == 0 {
		return diag.Errorf("No ccloud_automation_v1 found")
	}

	if len(automations) > 1 {
		return diag.Errorf("More than one ccloud_automation_v1 found (%d)", len(automations))
	}

	automation := automations[0]

	log.Printf("[DEBUG] Retrieved %s ccloud_automation_v1: %+v", automation.ID, automation)
	d.SetId(automation.ID)
	d.Set("name", automation.Name)
	d.Set("repository", automation.Repository)
	d.Set("repository_revision", automation.RepositoryRevision)
	d.Set("repository_authentication_enabled", automation.RepositoryAuthenticationEnabled)
	d.Set("project_id", automation.ProjectID)
	d.Set("timeout", automation.Timeout)
	d.Set("tags", automation.Tags)
	d.Set("created_at", automation.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", automation.UpdatedAt.Format(time.RFC3339))
	d.Set("type", automation.Type)
	d.Set("run_list", automation.RunList)

	chefAttributes, err := json.Marshal(automation.ChefAttributes)
	if err != nil {
		log.Printf("[DEBUG] dataSourceCCloudAutomationV1Read: Cannot marshal automation.ChefAttributes: %s", err)
	}
	d.Set("chef_attributes", string(chefAttributes))

	d.Set("log_level", automation.LogLevel)
	d.Set("debug", automation.Debug)
	d.Set("chef_version", automation.ChefVersion)
	d.Set("path", automation.Path)
	d.Set("arguments", automation.Arguments)
	d.Set("environment", automation.Environment)

	d.Set("region", GetRegion(d, config))

	return nil
}
