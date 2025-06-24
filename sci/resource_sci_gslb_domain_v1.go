package sci

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/andromeda/client/domains"
	"github.com/sapcc/andromeda/models"
)

func resourceSCIGSLBDomainV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIGSLBDomainV1Create,
		ReadContext:   resourceSCIGSLBDomainV1Read,
		UpdateContext: resourceSCIGSLBDomainV1Update,
		DeleteContext: resourceSCIGSLBDomainV1Delete,
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
			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"aliases": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"fqdn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mode": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"ROUND_ROBIN", "WEIGHTED", "GEOGRAPHIC", "AVAILABILITY",
				}, false),
				Optional: true,
				Default:  "ROUND_ROBIN",
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pools": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"service_provider": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"akamai", "f5",
				}, false),
				Optional: true,
				Default:  "akamai",
			},
			"record_type": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"A", "AAAA", "CNAME", "MX",
				}, false),
				Optional: true,
				Default:  "A",
			},

			// computed
			"cname_target": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provisioning_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
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
		},
	}
}

func resourceSCIGSLBDomainV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Domains

	// Create the domain
	adminStateUp := d.Get("admin_state_up").(bool)
	fqdn := strfmt.Hostname(d.Get("fqdn").(string))
	provider := d.Get("service_provider").(string)
	domain := &models.Domain{
		AdminStateUp: &adminStateUp,
		Fqdn:         &fqdn,
		Provider:     &provider,
	}

	if v, ok := d.GetOk("name"); ok && v != "" {
		domain.Name = ptr(v.(string))
	}
	if v, ok := d.GetOk("project_id"); ok && v != "" {
		domain.ProjectID = ptr(v.(string))
	}
	if v, ok := d.GetOk("pools"); ok {
		domain.Pools = expandToStrFmtUUIDSlice(v.([]any))
	}
	if v, ok := d.GetOk("mode"); ok && v != "" {
		domain.Mode = ptr(v.(string))
	}
	if v, ok := d.GetOk("record_type"); ok && v != "" {
		domain.RecordType = ptr(v.(string))
	}
	if v, ok := d.GetOk("aliases"); ok {
		domain.Aliases = expandToStringSlice(v.([]any))
	}

	opts := &domains.PostDomainsParams{
		Domain: domains.PostDomainsBody{
			Domain: domain,
		},
		Context: ctx,
	}
	res, err := client.PostDomains(opts)
	if err != nil {
		return diag.Errorf("error creating Andromeda domain: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Domain == nil {
		return diag.Errorf("error creating Andromeda domain: empty response")
	}

	log.Printf("[DEBUG] Created Andromeda domain: %v", res)

	id := string(res.Payload.Domain.ID)
	d.SetId(id)

	// waiting for the ACTIVE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := models.DomainProvisioningStatusACTIVE
	pending := models.DomainProvisioningStatusPENDINGCREATE
	domain, err = andromedaWaitForDomain(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetDomainResource(d, config, domain)

	return nil
}

func resourceSCIGSLBDomainV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Domains

	id := d.Id()
	domain, err := andromedaGetDomain(ctx, client, id)
	if err != nil {
		if _, ok := err.(*domains.GetDomainsDomainIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	andromedaSetDomainResource(d, config, domain)

	return nil
}

func resourceSCIGSLBDomainV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Domains

	id := d.Id()
	domain := &models.Domain{
		Fqdn:     ptr(strfmt.Hostname(d.Get("fqdn").(string))),
		Provider: ptr(d.Get("service_provider").(string)),
	}

	if d.HasChange("admin_state_up") {
		v := d.Get("admin_state_up").(bool)
		domain.AdminStateUp = &v
	}
	if d.HasChange("aliases") {
		v := d.Get("aliases").([]any)
		domain.Aliases = expandToStringSlice(v)
	}
	if d.HasChange("mode") {
		v := d.Get("mode").(string)
		domain.Mode = &v
	}
	if d.HasChange("name") {
		v := d.Get("name").(string)
		domain.Name = &v
	}
	if d.HasChange("pools") {
		v := d.Get("pools").([]any)
		domain.Pools = expandToStrFmtUUIDSlice(v)
	}
	if d.HasChange("project_id") {
		v := d.Get("project_id").(string)
		domain.ProjectID = &v
	}
	if d.HasChange("record_type") {
		v := d.Get("record_type").(string)
		domain.RecordType = &v
	}

	opts := &domains.PutDomainsDomainIDParams{
		Domain: domains.PutDomainsDomainIDBody{
			Domain: domain,
		},
		DomainID: strfmt.UUID(id),
		Context:  ctx,
	}
	_, err = client.PutDomainsDomainID(opts)
	if err != nil {
		return diag.Errorf("error updating Andromeda domain: %s", err)
	}

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutUpdate)
	target := models.DomainProvisioningStatusACTIVE
	pending := models.DomainProvisioningStatusPENDINGUPDATE
	domain, err = andromedaWaitForDomain(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetDomainResource(d, config, domain)

	return nil
}

func resourceSCIGSLBDomainV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Domains

	id := d.Id()
	opts := &domains.DeleteDomainsDomainIDParams{
		DomainID: strfmt.UUID(id),
		Context:  ctx,
	}
	_, err = client.DeleteDomainsDomainID(opts)
	if err != nil {
		if _, ok := err.(*domains.DeleteDomainsDomainIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Andromeda domain: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := "DELETED"
	pending := models.DomainProvisioningStatusPENDINGDELETE
	_, err = andromedaWaitForDomain(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func andromedaWaitForDomain(ctx context.Context, client domains.ClientService, id, target, pending string, timeout time.Duration) (*models.Domain, error) {
	log.Printf("[DEBUG] Waiting for %s domain to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    []string{pending},
		Refresh:    andromedaGetDomainStatus(ctx, client, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	domain, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*domains.GetDomainsDomainIDNotFound); ok && target == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s domain to become %s: %s", id, target, err)
	}

	return domain.(*models.Domain), nil
}

func andromedaGetDomainStatus(ctx context.Context, client domains.ClientService, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		domain, err := andromedaGetDomain(ctx, client, id)
		if err != nil {
			return nil, "", err
		}

		return domain, domain.ProvisioningStatus, nil
	}
}

func andromedaGetDomain(ctx context.Context, client domains.ClientService, id string) (*models.Domain, error) {
	opts := &domains.GetDomainsDomainIDParams{
		DomainID: strfmt.UUID(id),
		Context:  ctx,
	}
	res, err := client.GetDomainsDomainID(opts)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil || res.Payload.Domain == nil {
		return nil, fmt.Errorf("error reading Andromeda domain: empty response")
	}

	return res.Payload.Domain, nil
}

func andromedaSetDomainResource(d *schema.ResourceData, config *Config, domain *models.Domain) {
	_ = d.Set("admin_state_up", ptrValue(domain.AdminStateUp))
	_ = d.Set("aliases", domain.Aliases)
	_ = d.Set("fqdn", ptrValue(domain.Fqdn))
	_ = d.Set("mode", ptrValue(domain.Mode))
	_ = d.Set("name", ptrValue(domain.Name))
	_ = d.Set("pools", domain.Pools)
	_ = d.Set("project_id", ptrValue(domain.ProjectID))
	_ = d.Set("service_provider", ptrValue(domain.Provider))
	_ = d.Set("record_type", ptrValue(domain.RecordType))

	// computed
	_ = d.Set("cname_target", ptrValue(domain.CnameTarget))
	_ = d.Set("provisioning_status", domain.ProvisioningStatus)
	_ = d.Set("status", domain.Status)
	_ = d.Set("created_at", domain.CreatedAt.String())
	_ = d.Set("updated_at", domain.UpdatedAt.String())

	_ = d.Set("region", GetRegion(d, config))
}
