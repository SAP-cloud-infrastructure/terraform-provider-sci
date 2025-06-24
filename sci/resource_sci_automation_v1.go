package sci

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

func resourceSCIAutomationV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIAutomationV1Create,
		ReadContext:   resourceSCIAutomationV1Read,
		UpdateContext: resourceSCIAutomationV1Update,
		DeleteContext: resourceSCIAutomationV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"repository": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateURL,
			},

			"repository_revision": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "master",
				ValidateFunc: validation.NoZeroValues,
			},

			"repository_credentials": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Script", "Chef",
				}, false),
			},

			"tags": {
				Type:       schema.TypeMap,
				Optional:   true,
				Deprecated: "This field is not supported by the Lyra API",
			},

			// Chef parameters
			"run_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"path", "arguments", "environment"},
			},

			"chef_attributes": {
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{"path", "arguments", "environment"},
				ValidateFunc:     validateJSONObject,
				DiffSuppressFunc: diffSuppressJSONObject,
				StateFunc:        normalizeJSONString,
			},

			"log_level": {
				Type:          schema.TypeString,
				Optional:      true,
				Deprecated:    "This field is not supported by the Lyra API",
				ConflictsWith: []string{"path", "arguments", "environment"},
			},

			"debug": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"path", "arguments", "environment"},
			},

			"chef_version": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"path", "arguments", "environment"},
			},

			// Script parameters
			"path": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"run_list", "chef_attributes", "log_level", "debug", "chef_version"},
				ValidateFunc:  validation.NoZeroValues,
			},

			"arguments": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"run_list", "chef_attributes", "log_level", "debug", "chef_version"},
			},

			"environment": {
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"run_list", "chef_attributes", "log_level", "debug", "chef_version"},
			},

			// Computed
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

func resourceSCIAutomationV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Automation client: %s", err)
	}

	var chefAttributes map[string]any

	// Convert raw string into the map
	chefAttributesJSON := d.Get("chef_attributes").(string)
	if len(chefAttributesJSON) > 0 {
		err := json.Unmarshal([]byte(chefAttributesJSON), &chefAttributes)
		if err != nil {
			return diag.Errorf("Failed to unmarshal the JSON: %s", err)
		}
	}

	runList := d.Get("run_list").([]any)
	arguments := d.Get("arguments").([]any)
	environment := d.Get("environment").(map[string]any)
	tags := d.Get("tags").(map[string]any)

	createOpts := automations.CreateOpts{
		Name:                  d.Get("name").(string),
		Repository:            d.Get("repository").(string),
		RepositoryRevision:    d.Get("repository_revision").(string),
		RepositoryCredentials: d.Get("repository_credentials").(string),
		Timeout:               d.Get("timeout").(int),
		Tags:                  expandToMapStringString(tags),
		Type:                  d.Get("type").(string),
		// Chef
		RunList:        expandToStringSlice(runList),
		ChefAttributes: chefAttributes,
		LogLevel:       d.Get("log_level").(string),
		Debug:          d.Get("debug").(bool),
		ChefVersion:    d.Get("chef_version").(string),
		// Script
		Path:        d.Get("path").(string),
		Arguments:   expandToStringSlice(arguments),
		Environment: expandToMapStringString(environment),
	}

	log.Printf("[DEBUG] sci_automation_v1 create options: %#v", createOpts)

	automation, err := automations.Create(ctx, automationClient, createOpts).Extract()
	if err != nil {
		return diag.Errorf("Error creating sci_automation_v1: %s", err)
	}

	d.SetId(automation.ID)

	return resourceSCIAutomationV1Read(ctx, d, meta)
}

func resourceSCIAutomationV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	automation, err := automations.Get(ctx, automationClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Unable to retrieve sci_automation_v1"))
	}

	_ = d.Set("name", automation.Name)
	_ = d.Set("repository", automation.Repository)
	_ = d.Set("repository_revision", automation.RepositoryRevision)
	_ = d.Set("repository_authentication_enabled", automation.RepositoryAuthenticationEnabled)
	_ = d.Set("project_id", automation.ProjectID)
	_ = d.Set("timeout", automation.Timeout)
	_ = d.Set("tags", automation.Tags)
	_ = d.Set("created_at", automation.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", automation.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("type", automation.Type)
	_ = d.Set("run_list", automation.RunList)

	chefAttributes, err := json.Marshal(automation.ChefAttributes)
	if err != nil {
		log.Printf("[DEBUG] resourceSCIAutomationV1Read: Cannot marshal automation.ChefAttributes: %s", err)
	}
	_ = d.Set("chef_attributes", string(chefAttributes))

	_ = d.Set("log_level", automation.LogLevel)
	_ = d.Set("debug", automation.Debug)
	_ = d.Set("chef_version", automation.ChefVersion)
	_ = d.Set("path", automation.Path)
	_ = d.Set("arguments", automation.Arguments)
	_ = d.Set("environment", automation.Environment)

	_ = d.Set("region", GetRegion(d, config))

	return nil
}

func resourceSCIAutomationV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	var updateOpts automations.UpdateOpts

	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}

	if d.HasChange("repository") {
		updateOpts.Repository = d.Get("repository").(string)
	}

	if d.HasChange("repository_revision") {
		repositoryRevision := d.Get("repository_revision").(string)
		updateOpts.RepositoryRevision = &repositoryRevision
	}

	if d.HasChange("repository_credentials") {
		repositoryCredentials := d.Get("repository_credentials").(string)
		updateOpts.RepositoryCredentials = &repositoryCredentials
	}

	if d.HasChange("timeout") {
		updateOpts.Timeout = d.Get("timeout").(int)
	}

	if d.HasChange("tags") {
		tags := d.Get("tags").(map[string]any)
		updateOpts.Tags = expandToMapStringString(tags)
	}

	if d.HasChange("run_list") {
		runList := d.Get("run_list").([]any)
		updateOpts.RunList = expandToStringSlice(runList)
	}

	if d.HasChange("chef_attributes") {
		var chefAttributes map[string]any
		// Convert raw string into the map
		chefAttributesJSON := d.Get("chef_attributes").(string)
		if len(chefAttributesJSON) > 0 {
			err := json.Unmarshal([]byte(chefAttributesJSON), &chefAttributes)
			if err != nil {
				return diag.Errorf("Failed to unmarshal the JSON: %s", err)
			}
		}

		updateOpts.ChefAttributes = chefAttributes
	}

	if d.HasChange("log_level") {
		logLevel := d.Get("log_level").(string)
		updateOpts.LogLevel = &logLevel
	}

	if d.HasChange("debug") {
		debug := d.Get("debug").(bool)
		updateOpts.Debug = &debug
	}

	if d.HasChange("chef_version") {
		chefVersion := d.Get("chef_version").(string)
		updateOpts.ChefVersion = &chefVersion
	}

	if d.HasChange("path") {
		path := d.Get("path").(string)
		updateOpts.Path = &path
	}

	if d.HasChange("arguments") {
		arguments := d.Get("arguments").([]any)
		updateOpts.Arguments = expandToStringSlice(arguments)
	}

	if d.HasChange("environment") {
		environment := d.Get("environment").(map[string]any)
		updateOpts.Environment = expandToMapStringString(environment)
	}

	_, err = automations.Update(ctx, automationClient, d.Id(), updateOpts).Extract()
	if err != nil {
		return diag.Errorf("Error updating sci_automation_v1: %s", err)
	}

	return resourceSCIAutomationV1Read(ctx, d, meta)
}

func resourceSCIAutomationV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	log.Printf("[DEBUG] Deleting sci_automation_v1: %s", d.Id())
	err = automations.Delete(ctx, automationClient, d.Id()).ExtractErr()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Error deleting sci_automation_v1"))
	}

	return nil
}
