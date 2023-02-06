package ccloud

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/billing/masterdata/projects"
)

func dataSourceCCloudBillingProjectMasterdata() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCCloudBillingProjectMasterdataRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"project_name": {
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

			"parent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"project_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"revenue_relevance": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"business_criticality": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"number_of_endusers": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"additional_information": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_primary_contact_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_primary_contact_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_operator_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_operator_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_security_expert_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_security_expert_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_product_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_product_owner_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_controller_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_controller_email": {
				Type:     schema.TypeString,
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
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inherited": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

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

func dataSourceCCloudBillingProjectMasterdataRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack billing client: %s", err)
	}

	projectID := d.Get("project_id").(string)

	var project *projects.Project
	if projectID == "" {
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
		project, err = projects.Get(billing, projectID).Extract()
		if err != nil {
			return fmt.Errorf("Error getting billing project masterdata: %s", err)
		}
	}

	log.Printf("[DEBUG] Retrieved project masterdata: %+v", project)

	d.SetId(project.ProjectID)

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
