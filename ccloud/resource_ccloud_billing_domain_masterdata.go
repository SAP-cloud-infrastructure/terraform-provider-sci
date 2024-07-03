package ccloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/billing/masterdata/domains"

	"github.com/gophercloud/gophercloud"
)

func resourceCCloudBillingDomainMasterdata() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceCCloudBillingDomainMasterdataRead,
		UpdateContext: resourceCCloudBillingDomainMasterdataCreateOrUpdate,
		CreateContext: resourceCCloudBillingDomainMasterdataCreateOrUpdate,
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

			"domain_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"domain_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"description": {
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

			"cost_object": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"projects_can_inherit": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"IO", "CC", "WBS", "SO",
							}, false),
						},
					},
				},
			},

			"collector": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
		},
	}
}

func resourceCCloudBillingDomainMasterdataCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack billing client: %s", err)
	}

	domainID := d.Get("domain_id").(string)
	if d.Id() == "" && domainID == "" {
		// first call, expecting to get current scope domain
		identityClient, err := config.IdentityV3Client(GetRegion(d, config))
		if err != nil {
			return diag.Errorf("Error creating OpenStack identity client: %s", err)
		}

		tokenDetails, err := getTokenDetails(identityClient)
		if err != nil {
			return diag.FromErr(err)
		}

		if tokenDetails.domain == nil {
			return diag.Errorf("Error getting billing domain scope: %s", err)
		}

		domainID = tokenDetails.domain.ID
	}

	domain, err := domains.Get(billing, domainID).Extract()
	if err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); d.Id() != "" || !ok {
			return diag.Errorf("Error getting billing domain masterdata: %s", err)
		}
		log.Printf("[DEBUG] Error getting billing domain masterdata, probably this domain was not created yet: %s", err)
		domain = &domains.Domain{DomainID: domainID}
	}

	log.Printf("[DEBUG] Retrieved domain masterdata before the created/update: %+v", domain)

	// API doesn't support partial update, thus prefilling the update options with the existing data
	opts := domains.DomainToUpdateOpts(domain)
	opts.DomainID = replaceEmptyString(d, "domain_id", opts.DomainID)
	opts.DomainName = replaceEmptyString(d, "domain_name", opts.DomainName)
	opts.ResponsiblePrimaryContactID = replaceEmptyString(d, "responsible_primary_contact_id", opts.ResponsiblePrimaryContactID)
	opts.ResponsiblePrimaryContactEmail = replaceEmptyString(d, "responsible_primary_contact_email", opts.ResponsiblePrimaryContactEmail)
	opts.AdditionalInformation = replaceEmptyString(d, "additional_information", opts.AdditionalInformation)
	opts.Collector = replaceEmptyString(d, "collector", opts.Collector)

	if v := billingDomainExpandCostObject(d.Get("cost_object")); v != (domains.CostObject{}) {
		opts.CostObject = v
	}

	log.Printf("[DEBUG] Updating %s domain masterdata: %+v", opts.DomainID, opts)

	_, err = domains.Update(billing, opts.DomainID, opts).Extract()
	if err != nil {
		return diag.Errorf("Error updating billing domain masterdata: %s", err)
	}

	if d.Id() == "" {
		d.SetId(opts.DomainID)
	}

	return resourceCCloudBillingDomainMasterdataRead(ctx, d, meta)
}

func resourceCCloudBillingDomainMasterdataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack billing client: %s", err)
	}

	domain, err := domains.Get(billing, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Error getting billing domain masterdata"))
	}

	log.Printf("[DEBUG] Retrieved domain masterdata: %+v", domain)

	d.Set("domain_id", domain.DomainID)
	d.Set("domain_name", domain.DomainName)
	d.Set("description", domain.Description)
	d.Set("responsible_primary_contact_id", domain.ResponsiblePrimaryContactID)
	d.Set("responsible_primary_contact_email", domain.ResponsiblePrimaryContactEmail)
	d.Set("additional_information", domain.AdditionalInformation)
	d.Set("cost_object", billingDomainFlattenCostObject(domain.CostObject))
	d.Set("created_at", domain.CreatedAt.Format(time.RFC3339))
	d.Set("changed_at", domain.ChangedAt.Format(time.RFC3339))
	d.Set("changed_by", domain.ChangedBy)
	d.Set("is_complete", domain.IsComplete)
	d.Set("missing_attributes", domain.MissingAttributes)
	d.Set("collector", domain.Collector)

	d.Set("region", GetRegion(d, config))

	return nil
}
