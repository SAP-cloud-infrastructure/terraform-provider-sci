package sci

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/v2/billing/masterdata/projects"
)

func resourceSCIBillingProjectMasterdata() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSCIBillingProjectMasterdataRead,
		UpdateContext: resourceSCIBillingProjectMasterdataCreateOrUpdate,
		CreateContext: resourceSCIBillingProjectMasterdataCreateOrUpdate,
		Delete:        schema.RemoveFromState,
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
			},

			"revenue_relevance": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"generating",
					"enabling",
					"other",
				}, false),
			},

			"business_criticality": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"dev",
					"test",
					"prod",
				}, false),
			},

			"number_of_endusers": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(-1),
			},

			"customer": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"additional_information": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"responsible_primary_contact_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"responsible_primary_contact_email": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"responsible_operator_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"responsible_operator_email": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"responsible_inventory_role_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"responsible_inventory_role_email": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"responsible_infrastructure_coordinator_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"responsible_infrastructure_coordinator_email": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"cost_object": {
				Type:     schema.TypeList,
				Optional: true,
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

			"environment": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Prod",
					"QA",
					"Admin",
					"DEV",
					"Demo",
					"Train",
					"Sandbox",
					"Lab",
					"Test",
				}, false),
			},

			"soft_license_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Revenue Generating",
					"Training & Demo",
					"Development",
					"Test & QS",
					"Administration",
					"Make",
					"Virtualization-Host",
					"Productive",
				}, false),
			},

			"type_of_data": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"SAP Business Process",
					"Customer Cloud Service",
					"Customer Business Process",
					"Training & Demo Cloud",
				}, false),
			},

			"gpu_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"contains_pii_dpp_hr": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"contains_external_customer_data": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"ext_certification": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"c5": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"iso": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"pci": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"soc1": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"soc2": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"sox": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
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

func resourceSCIBillingProjectMasterdataCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	billing, err := config.billingClient(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack billing client: %s", err)
	}

	projectID := d.Get("project_id").(string)
	if d.Id() == "" && projectID == "" {
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
		if d.Id() != "" || !gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			return diag.Errorf("Error getting billing project masterdata: %s", err)
		}
		log.Printf("[DEBUG] Error getting billing project masterdata, probably this project was not created yet: %s", err)
		project = &projects.Project{ProjectID: projectID}
	}

	log.Printf("[DEBUG] Retrieved project masterdata before the created/update: %+v", project)

	// API doesn't support partial update, thus prefilling the update options with the existing data
	opts := projects.ProjectToUpdateOpts(project)
	opts.ResponsiblePrimaryContactID = replaceEmptyString(d, "responsible_primary_contact_id", opts.ResponsiblePrimaryContactID)
	opts.ResponsiblePrimaryContactEmail = replaceEmptyString(d, "responsible_primary_contact_email", opts.ResponsiblePrimaryContactEmail)
	opts.ResponsibleOperatorID = replaceEmptyString(d, "responsible_operator_id", opts.ResponsibleOperatorID)
	opts.ResponsibleOperatorEmail = replaceEmptyString(d, "responsible_operator_email", opts.ResponsibleOperatorEmail)
	opts.ResponsibleInventoryRoleID = replaceEmptyString(d, "responsible_inventory_role_id", opts.ResponsibleInventoryRoleID)
	opts.ResponsibleInventoryRoleEmail = replaceEmptyString(d, "responsible_inventory_role_email", opts.ResponsibleInventoryRoleEmail)
	opts.ResponsibleInfrastructureCoordinatorID = replaceEmptyString(d, "responsible_infrastructure_coordinator_id", opts.ResponsibleInfrastructureCoordinatorID)
	opts.ResponsibleInfrastructureCoordinatorEmail = replaceEmptyString(d, "responsible_infrastructure_coordinator_email", opts.ResponsibleInfrastructureCoordinatorEmail)
	opts.Customer = replaceEmptyString(d, "customer", opts.Customer)
	opts.Environment = replaceEmptyString(d, "environment", opts.Environment)
	opts.SoftLicenseMode = replaceEmptyString(d, "soft_license_mode", opts.SoftLicenseMode)
	opts.TypeOfData = replaceEmptyString(d, "type_of_data", opts.TypeOfData)
	opts.RevenueRelevance = replaceEmptyString(d, "revenue_relevance", opts.RevenueRelevance)
	opts.BusinessCriticality = replaceEmptyString(d, "business_criticality", opts.BusinessCriticality)
	opts.AdditionalInformation = replaceEmptyString(d, "additional_information", opts.AdditionalInformation)
	opts.GPUEnabled = replaceEmptyBool(d, "gpu_enabled", opts.GPUEnabled)
	opts.ContainsPIIDPPHR = replaceEmptyBool(d, "contains_pii_dpp_hr", opts.ContainsPIIDPPHR)
	opts.ContainsExternalCustomerData = replaceEmptyBool(d, "contains_external_customer_data", opts.ContainsExternalCustomerData)
	opts.ExtCertification = billingProjectExpandExtCertificationV1(d.Get("ext_certification"))

	if v, ok := getOkExists(d, "number_of_endusers"); ok {
		opts.NumberOfEndusers = v.(int)
	}

	if v := billingProjectExpandCostObject(d.Get("cost_object")); v != (projects.CostObject{}) {
		opts.CostObject = v
	}

	// admin only parameters
	opts.ProjectID = replaceEmptyString(d, "project_id", opts.ProjectID)
	opts.ProjectName = replaceEmptyString(d, "project_name", opts.ProjectName)
	opts.DomainID = replaceEmptyString(d, "domain_id", opts.DomainID)
	opts.DomainName = replaceEmptyString(d, "domain_name", opts.DomainName)
	opts.ParentID = replaceEmptyString(d, "parent_id", opts.ParentID)
	opts.ProjectType = replaceEmptyString(d, "project_type", opts.ProjectType)

	log.Printf("[DEBUG] Updating %s project masterdata: %+v", opts.ProjectID, opts)

	_, err = projects.Update(ctx, billing, opts.ProjectID, opts).Extract()
	if err != nil {
		return diag.Errorf("Error updating billing project masterdata: %s", err)
	}

	if d.Id() == "" {
		d.SetId(opts.ProjectID)
	}

	return resourceSCIBillingProjectMasterdataRead(ctx, d, meta)
}

func resourceSCIBillingProjectMasterdataRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	billing, err := config.billingClient(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack billing client: %s", err)
	}

	project, err := projects.Get(ctx, billing, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Error getting billing project masterdata"))
	}

	log.Printf("[DEBUG] Retrieved project masterdata: %+v", project)

	_ = d.Set("project_id", project.ProjectID)
	_ = d.Set("project_name", project.ProjectName)
	_ = d.Set("domain_id", project.DomainID)
	_ = d.Set("domain_name", project.DomainName)
	_ = d.Set("description", project.Description)
	_ = d.Set("parent_id", project.ParentID)
	_ = d.Set("project_type", project.ProjectType)
	_ = d.Set("responsible_primary_contact_id", project.ResponsiblePrimaryContactID)
	_ = d.Set("responsible_primary_contact_email", project.ResponsiblePrimaryContactEmail)
	_ = d.Set("responsible_operator_id", project.ResponsibleOperatorID)
	_ = d.Set("responsible_operator_email", project.ResponsibleOperatorEmail)
	_ = d.Set("responsible_inventory_role_id", project.ResponsibleInventoryRoleID)
	_ = d.Set("responsible_inventory_role_email", project.ResponsibleInventoryRoleEmail)
	_ = d.Set("responsible_infrastructure_coordinator_id", project.ResponsibleInfrastructureCoordinatorID)
	_ = d.Set("responsible_infrastructure_coordinator_email", project.ResponsibleInfrastructureCoordinatorEmail)
	_ = d.Set("customer", project.Customer)
	_ = d.Set("environment", project.Environment)
	_ = d.Set("soft_license_mode", project.SoftLicenseMode)
	_ = d.Set("type_of_data", project.TypeOfData)
	_ = d.Set("gpu_enabled", project.GPUEnabled)
	_ = d.Set("contains_pii_dpp_hr", project.ContainsPIIDPPHR)
	_ = d.Set("contains_external_customer_data", project.ContainsExternalCustomerData)
	_ = d.Set("ext_certification", billingProjectFlattenExtCertificationV1(project.ExtCertification))
	_ = d.Set("revenue_relevance", project.RevenueRelevance)
	_ = d.Set("business_criticality", project.BusinessCriticality)
	_ = d.Set("number_of_endusers", project.NumberOfEndusers)
	_ = d.Set("additional_information", project.AdditionalInformation)
	_ = d.Set("cost_object", billingProjectFlattenCostObject(project.CostObject))
	_ = d.Set("created_at", project.CreatedAt.Format(time.RFC3339))
	_ = d.Set("changed_at", project.ChangedAt.Format(time.RFC3339))
	_ = d.Set("changed_by", project.ChangedBy)
	_ = d.Set("is_complete", project.IsComplete)
	_ = d.Set("missing_attributes", project.MissingAttributes)
	_ = d.Set("collector", project.Collector)

	_ = d.Set("region", GetRegion(d, config))

	return nil
}
