package sci

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/andromeda/client/administrative"
	"github.com/sapcc/andromeda/models"
)

func resourceSCIGSLBQuotaV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIGSLBQuotaV1Create,
		ReadContext:   resourceSCIGSLBQuotaV1Read,
		UpdateContext: resourceSCIGSLBQuotaV1Update,
		DeleteContext: resourceSCIGSLBQuotaV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceSCIGSLBQuotaV1V0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceSCIGSLBQuotaV1StateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"datacenter": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"domain": {
				Type:       schema.TypeInt,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use domain_akamai and domain_f5 instead.",
			},
			"domain_akamai": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"domain_f5": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"member": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"monitor": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"pool": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			// computed
			"in_use_datacenter": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_domain": {
				Type:       schema.TypeInt,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use in_use_domain_akamai and in_use_domain_f5 instead.",
			},
			"in_use_domain_akamai": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_domain_f5": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_member": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_monitor": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_pool": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceSCIGSLBQuotaV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Administrative

	projectID := d.Get("project_id").(string)

	quota := &models.Quota{}
	if v, ok := getOkExists(d, "datacenter"); ok && v != "" {
		quota.Datacenter = ptr(int64(v.(int)))
	}
	if v, ok := getOkExists(d, "domain_akamai"); ok && v != "" {
		quota.DomainAkamai = ptr(int64(v.(int)))
	}
	if v, ok := getOkExists(d, "domain_f5"); ok && v != "" {
		quota.DomainF5 = ptr(int64(v.(int)))
	}
	if v, ok := getOkExists(d, "member"); ok && v != "" {
		quota.Member = ptr(int64(v.(int)))
	}
	if v, ok := getOkExists(d, "monitor"); ok && v != "" {
		quota.Monitor = ptr(int64(v.(int)))
	}
	if v, ok := getOkExists(d, "pool"); ok && v != "" {
		quota.Pool = ptr(int64(v.(int)))
	}

	opts := &administrative.PutQuotasProjectIDParams{
		Quota: administrative.PutQuotasProjectIDBody{
			Quota: quota,
		},
		ProjectID: projectID,
		Context:   ctx,
	}
	res, err := client.PutQuotasProjectID(opts)
	if err != nil {
		return diag.Errorf("error creating Andromeda quota: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Quota == nil {
		return diag.Errorf("error creating Andromeda quota: empty response")
	}

	log.Printf("[DEBUG] Created Andromeda quota: %v", res)

	d.SetId(projectID)

	return resourceSCIGSLBQuotaV1Read(ctx, d, meta)
}

func resourceSCIGSLBQuotaV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Administrative

	id := d.Id()
	opts := &administrative.GetQuotasProjectIDParams{
		ProjectID: id,
		Context:   ctx,
	}
	res, err := client.GetQuotasProjectID(opts)
	if err != nil {
		return diag.Errorf("error reading Andromeda quota: %s, %T", err, err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error reading Andromeda quota: empty response")
	}
	if err != nil {
		if _, ok := err.(*administrative.GetQuotasProjectIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	andromedaSetQuotaResource(d, config, res.Payload)

	return nil
}

func resourceSCIGSLBQuotaV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Administrative

	id := d.Id()
	quota := &models.Quota{}

	if d.HasChange("datacenter") {
		v := d.Get("datacenter").(int)
		quota.Datacenter = ptr(int64(v))
	}
	if d.HasChange("domain_akamai") {
		v := d.Get("domain_akamai").(int)
		quota.DomainAkamai = ptr(int64(v))
	}
	if d.HasChange("domain_f5") {
		v := d.Get("domain_f5").(int)
		quota.DomainF5 = ptr(int64(v))
	}
	if d.HasChange("member") {
		v := d.Get("member").(int)
		quota.Member = ptr(int64(v))
	}
	if d.HasChange("monitor") {
		v := d.Get("monitor").(int)
		quota.Monitor = ptr(int64(v))
	}
	if d.HasChange("pool") {
		v := d.Get("pool").(int)
		quota.Pool = ptr(int64(v))
	}

	opts := &administrative.PutQuotasProjectIDParams{
		Quota: administrative.PutQuotasProjectIDBody{
			Quota: quota,
		},
		ProjectID: id,
		Context:   ctx,
	}
	res, err := client.PutQuotasProjectID(opts)
	if err != nil {
		return diag.Errorf("error updating Andromeda quota: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Quota == nil {
		return diag.Errorf("error updating Andromeda quota: empty response")
	}

	return resourceSCIGSLBQuotaV1Read(ctx, d, meta)
}

func resourceSCIGSLBQuotaV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Administrative

	id := d.Id()
	opts := &administrative.DeleteQuotasProjectIDParams{
		ProjectID: id,
		Context:   ctx,
	}
	_, err = client.DeleteQuotasProjectID(opts)
	if err != nil {
		if _, ok := err.(*administrative.DeleteQuotasProjectIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Andromeda quota: %s", err)
	}

	return nil
}

func andromedaSetQuotaResource(d *schema.ResourceData, config *Config, q *administrative.GetQuotasProjectIDOKBody) {
	_ = d.Set("datacenter", ptrValue(q.Quota.Datacenter))
	_ = d.Set("domain_akamai", ptrValue(q.Quota.DomainAkamai))
	_ = d.Set("domain_f5", ptrValue(q.Quota.DomainF5))
	_ = d.Set("member", ptrValue(q.Quota.Member))
	_ = d.Set("monitor", ptrValue(q.Quota.Monitor))
	_ = d.Set("pool", ptrValue(q.Quota.Pool))

	// computed
	_ = d.Set("in_use_datacenter", q.Quota.InUseDatacenter)
	_ = d.Set("in_use_domain_akamai", q.Quota.InUseDomainAkamai)
	_ = d.Set("in_use_domain_f5", q.Quota.InUseDomainF5)
	_ = d.Set("in_use_member", q.Quota.InUseMember)
	_ = d.Set("in_use_monitor", q.Quota.InUseMonitor)
	_ = d.Set("in_use_pool", q.Quota.InUsePool)

	_ = d.Set("region", GetRegion(d, config))
}
