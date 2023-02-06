package ccloud

import (
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/billing/masterdata/projects"
)

func resourceCCloudBillingProjectMasterdata() *schema.Resource {
	return &schema.Resource{
		Read:   resourceCCloudBillingProjectMasterdataRead,
		Update: resourceCCloudBillingProjectMasterdataCreateOrUpdate,
		Create: resourceCCloudBillingProjectMasterdataCreateOrUpdate,
		Delete: schema.RemoveFromState,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// admin only parameters
			"project_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"domain_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"domain_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"parent_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"project_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// user parameters
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"revenue_relevance": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"generating", "enabling", "other",
				}, false),
			},

			"business_criticality": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"dev", "test", "prod",
				}, false),
			},

			"number_of_endusers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(-1),
			},

			"additional_information": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_primary_contact_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"responsible_primary_contact_email": {
				Type:     schema.TypeString,
				Required: true,
			},

			"responsible_operator_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_operator_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_security_expert_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_security_expert_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_product_owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_product_owner_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_controller_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_controller_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"inventory_role": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_inventory_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_inventory_role_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"infrastructure_coordinator": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_infrastructure_coordinator_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_infrastructure_coordinator_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cost_object": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inherited": {
							Type:          schema.TypeBool,
							Optional:      true,
							ConflictsWith: []string{"cost_object.0.name", "cost_object.0.type"},
						},
						"name": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"cost_object.0.inherited"},
						},
						"type": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"cost_object.0.inherited"},
							ValidateFunc: validation.StringInSlice([]string{
								"IO", "CC", "WBS", "SO",
							}, false),
						},
					},
				},
			},

			// computed parameters
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"changed_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"changed_by": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"is_complete": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"missing_attributes": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"collector": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCCloudBillingProjectMasterdataCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack billing client: %s", err)
	}

	projectID := d.Get("project_id").(string)

	var project *projects.Project
	if d.Id() == "" && projectID == "" {
		// first call, expecting to get current scope project
		allPages, err := projects.List(billing).AllPages()
		if err != nil {
			return fmt.Errorf("Error getting billing project masterdata: %s", err)
		}

		allProjects, err := projects.ExtractProjects(allPages)
		if err != nil {
			return fmt.Errorf("Error extracting billing projects masterdata: %s", err)
		}

		if len(allProjects) != 1 {
			return fmt.Errorf("Error getting billing project masterdata: expecting 1 project, got %d", len(allProjects))
		}

		project = &allProjects[0]
	} else {
		// admin mode, when the project doesn't correspond to the scope
		// or during the update, when project_id was already set
		project, err = projects.Get(billing, projectID).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); d.Id() != "" || !ok {
				return fmt.Errorf("Error getting billing project masterdata: %s", err)
			}
			log.Printf("[DEBUG] Error getting billing project masterdata, probably this project was not created yet: %s", err)
			project = &projects.Project{ProjectID: projectID}
		}
	}

	log.Printf("[DEBUG] Retrieved project masterdata before the created/update: %+v", project)

	// API doesn't support partial update, thus prefilling the update options with the existing data
	opts := projects.ProjectToUpdateOpts(project)
	opts.ResponsiblePrimaryContactID = replaceEmpty(d, "responsible_primary_contact_id", opts.ResponsiblePrimaryContactID)
	opts.ResponsiblePrimaryContactEmail = replaceEmpty(d, "responsible_primary_contact_email", opts.ResponsiblePrimaryContactEmail)
	opts.ResponsibleOperatorID = replaceEmpty(d, "responsible_operator_id", opts.ResponsibleOperatorID)
	opts.ResponsibleOperatorEmail = replaceEmpty(d, "responsible_operator_email", opts.ResponsibleOperatorEmail)
	opts.ResponsibleSecurityExpertID = replaceEmpty(d, "responsible_security_expert_id", opts.ResponsibleSecurityExpertID)
	opts.ResponsibleSecurityExpertEmail = replaceEmpty(d, "responsible_security_expert_email", opts.ResponsibleSecurityExpertEmail)
	opts.ResponsibleProductOwnerID = replaceEmpty(d, "responsible_product_owner_id", opts.ResponsibleProductOwnerID)
	opts.ResponsibleProductOwnerEmail = replaceEmpty(d, "responsible_product_owner_email", opts.ResponsibleProductOwnerEmail)
	opts.ResponsibleControllerID = replaceEmpty(d, "responsible_controller_id", opts.ResponsibleControllerID)
	opts.ResponsibleControllerEmail = replaceEmpty(d, "responsible_controller_email", opts.ResponsibleControllerEmail)
	opts.RevenueRelevance = replaceEmpty(d, "revenue_relevance", opts.RevenueRelevance)
	opts.BusinessCriticality = replaceEmpty(d, "business_criticality", opts.BusinessCriticality)
	opts.AdditionalInformation = replaceEmpty(d, "additional_information", opts.AdditionalInformation)

	if v, ok := d.GetOkExists("number_of_endusers"); ok {
		opts.NumberOfEndusers = v.(int)
	}

	if v := billingProjectExpandCostObject(d.Get("cost_object")); v != (projects.CostObject{}) {
		opts.CostObject = v
	}

	// admin only parameters
	opts.ProjectID = replaceEmpty(d, "project_id", opts.ProjectID)
	opts.ProjectName = replaceEmpty(d, "project_name", opts.ProjectName)
	opts.DomainID = replaceEmpty(d, "domain_id", opts.DomainID)
	opts.DomainName = replaceEmpty(d, "domain_name", opts.DomainName)
	opts.ParentID = replaceEmpty(d, "parent_id", opts.ParentID)
	opts.ProjectType = replaceEmpty(d, "project_type", opts.ProjectType)

	log.Printf("[QUOTA] Updating %s project masterdata: %+v", opts.ProjectID, opts)

	_, err = projects.Update(billing, opts.ProjectID, opts).Extract()
	if err != nil {
		return fmt.Errorf("Error updating billing project masterdata: %s", err)
	}

	if d.Id() == "" {
		d.SetId(opts.ProjectID)
	}

	return resourceCCloudBillingProjectMasterdataRead(d, meta)
}

func resourceCCloudBillingProjectMasterdataRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack billing client: %s", err)
	}

	project, err := projects.Get(billing, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Error getting billing project masterdata")
	}

	log.Printf("[DEBUG] Retrieved project masterdata: %+v", project)

	d.Set("project_id", project.ProjectID)
	d.Set("project_name", project.ProjectName)
	d.Set("domain_id", project.DomainID)
	d.Set("domain_name", project.DomainName)
	d.Set("description", project.Description)
	d.Set("parent_id", project.ParentID)
	d.Set("project_type", project.ProjectType)
	d.Set("responsible_primary_contact_id", project.ResponsiblePrimaryContactID)
	d.Set("responsible_primary_contact_email", project.ResponsiblePrimaryContactEmail)
	d.Set("responsible_operator_id", project.ResponsibleOperatorID)
	d.Set("responsible_operator_email", project.ResponsibleOperatorEmail)
	d.Set("responsible_security_expert_id", project.ResponsibleSecurityExpertID)
	d.Set("responsible_security_expert_email", project.ResponsibleSecurityExpertEmail)
	d.Set("responsible_product_owner_id", project.ResponsibleProductOwnerID)
	d.Set("responsible_product_owner_email", project.ResponsibleProductOwnerEmail)
	d.Set("responsible_controller_id", project.ResponsibleControllerID)
	d.Set("responsible_controller_email", project.ResponsibleControllerEmail)
	d.Set("revenue_relevance", project.RevenueRelevance)
	d.Set("business_criticality", project.BusinessCriticality)
	d.Set("number_of_endusers", project.NumberOfEndusers)
	d.Set("additional_information", project.AdditionalInformation)
	d.Set("cost_object", billingProjectFlattenCostObject(project.CostObject))
	d.Set("created_at", project.CreatedAt.Format(time.RFC3339))
	d.Set("changed_at", project.ChangedAt.Format(time.RFC3339))
	d.Set("changed_by", project.ChangedBy)
	d.Set("is_complete", project.IsComplete)
	d.Set("missing_attributes", project.MissingAttributes)
	d.Set("collector", project.Collector)

	d.Set("region", GetRegion(d, config))

	return nil
}
