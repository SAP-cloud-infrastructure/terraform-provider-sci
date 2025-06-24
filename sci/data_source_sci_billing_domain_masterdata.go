package sci

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/v2/billing/masterdata/domains"
)

func dataSourceSCIBillingDomainMasterdata() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIBillingDomainMasterdataRead,

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

func dataSourceSCIBillingDomainMasterdataRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	billing, err := config.billingClient(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack billing client: %s", err)
	}

	domainID := d.Get("domain_id").(string)
	if domainID == "" {
		// first call, expecting to get current scope domain
		identityClient, err := config.IdentityV3Client(ctx, GetRegion(d, config))
		if err != nil {
			return diag.Errorf("Error creating OpenStack identity client: %s", err)
		}

		tokenDetails, err := getTokenDetails(ctx, identityClient)
		if err != nil {
			return diag.FromErr(err)
		}

		if tokenDetails.domain == nil {
			return diag.Errorf("Error getting billing domain scope: %s", err)
		}

		domainID = tokenDetails.domain.ID
	}

	domain, err := domains.Get(ctx, billing, domainID).Extract()
	if err != nil {
		return diag.Errorf("Error getting billing domain masterdata: %s", err)
	}

	log.Printf("[DEBUG] Retrieved domain masterdata: %+v", domain)

	d.SetId(domain.DomainID)

	_ = d.Set("domain_id", domain.DomainID)
	_ = d.Set("domain_name", domain.DomainName)
	_ = d.Set("description", domain.Description)
	_ = d.Set("responsible_primary_contact_id", domain.ResponsiblePrimaryContactID)
	_ = d.Set("responsible_primary_contact_email", domain.ResponsiblePrimaryContactEmail)
	_ = d.Set("additional_information", domain.AdditionalInformation)
	_ = d.Set("cost_object", billingDomainFlattenCostObject(domain.CostObject))
	_ = d.Set("created_at", domain.CreatedAt.Format(time.RFC3339))
	_ = d.Set("changed_at", domain.ChangedAt.Format(time.RFC3339))
	_ = d.Set("changed_by", domain.ChangedBy)
	_ = d.Set("is_complete", domain.IsComplete)
	_ = d.Set("missing_attributes", domain.MissingAttributes)
	_ = d.Set("collector", domain.Collector)

	_ = d.Set("region", GetRegion(d, config))

	return nil
}
