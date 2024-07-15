package ccloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/v2/billing/masterdata/projects"
)

func dataSourceCCloudBillingProjectMasterdata() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCCloudBillingProjectMasterdataRead,

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

			"responsible_inventory_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_inventory_role_email": {
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

			"environment": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"soft_license_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"type_of_data": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"gpu_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"contains_pii_dpp_hr": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"contains_external_customer_data": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ext_certification": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"c5": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"iso": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"pci": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"soc1": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"soc2": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"sox": {
							Type:     schema.TypeBool,
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

func dataSourceCCloudBillingProjectMasterdataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	billing, err := config.billingClient(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack billing client: %s", err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		// first call, expecting to get current scope project
		identityClient, err := config.IdentityV3Client(ctx, GetRegion(d, config))
		if err != nil {
			return diag.Errorf("Error creating OpenStack identity client: %s", err)
		}

		tokenDetails, err := getTokenDetails(ctx, identityClient)
		if err != nil {
			return diag.FromErr(err)
		}

		if tokenDetails.project == nil {
			return diag.Errorf("Error getting billing project scope: %s", err)
		}

		projectID = tokenDetails.project.ID
	}

	project, err := projects.Get(ctx, billing, projectID).Extract()
	if err != nil {
		return diag.Errorf("Error getting billing project masterdata: %s", err)
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
	d.Set("responsible_inventory_role_id", project.ResponsibleInventoryRoleID)
	d.Set("responsible_inventory_role_email", project.ResponsibleInventoryRoleEmail)
	d.Set("responsible_infrastructure_coordinator_id", project.ResponsibleInfrastructureCoordinatorID)
	d.Set("responsible_infrastructure_coordinator_email", project.ResponsibleInfrastructureCoordinatorEmail)
	d.Set("customer", project.Customer)
	d.Set("environment", project.Environment)
	d.Set("soft_license_mode", project.SoftLicenseMode)
	d.Set("type_of_data", project.TypeOfData)
	d.Set("gpu_enabled", project.GPUEnabled)
	d.Set("contains_pii_dpp_hr", project.ContainsPIIDPPHR)
	d.Set("contains_external_customer_data", project.ContainsExternalCustomerData)
	d.Set("ext_certification", billingProjectFlattenExtCertificationV1(project.ExtCertification))
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
