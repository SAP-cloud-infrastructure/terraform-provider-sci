package ccloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/billing/masterdata/domains"
)

func dataSourceCCloudBillingDomainMasterdata() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCCloudBillingDomainMasterdataRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"domain_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
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

			"cost_object": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"projects_can_inherit": {
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

func dataSourceCCloudBillingDomainMasterdataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack billing client: %s", err)
	}

	domainID := d.Get("domain_id").(string)
	if domainID == "" {
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
		return diag.Errorf("Error getting billing domain masterdata: %s", err)
	}

	log.Printf("[DEBUG] Retrieved domain masterdata: %+v", domain)

	d.SetId(domain.DomainID)

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
