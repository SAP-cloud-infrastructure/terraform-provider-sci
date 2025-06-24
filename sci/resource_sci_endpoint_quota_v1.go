package sci

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/archer/client/quota"
	"github.com/sapcc/archer/models"
)

func resourceSCIEndpointQuotaV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIEndpointQuotaV1Create,
		ReadContext:   resourceSCIEndpointQuotaV1Read,
		UpdateContext: resourceSCIEndpointQuotaV1Update,
		DeleteContext: resourceSCIEndpointQuotaV1Delete,
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
			"endpoint": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"service": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			// computed
			"in_use_endpoint": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"in_use_service": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceSCIEndpointQuotaV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Quota

	projectID := d.Get("project_id").(string)
	req := &models.Quota{}
	if v, ok := d.GetOk("endpoint"); ok && v != 0 {
		req.Endpoint = int64(v.(int))
	}
	if v, ok := d.GetOk("service"); ok && v != 0 {
		req.Service = int64(v.(int))
	}

	opts := &quota.PutQuotasProjectIDParams{
		Body:      req,
		ProjectID: projectID,
		Context:   ctx,
	}
	res, err := client.PutQuotasProjectID(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error creating Archer quota: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error creating Archer quota: empty response")
	}

	log.Printf("[DEBUG] Created Archer quota: %v", res)

	d.SetId(projectID)

	return resourceSCIEndpointQuotaV1Read(ctx, d, meta)
}

func resourceSCIEndpointQuotaV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Quota

	id := d.Id()
	opts := &quota.GetQuotasProjectIDParams{
		ProjectID: id,
		Context:   ctx,
	}
	res, err := client.GetQuotasProjectID(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error reading Archer quota: %s, %T", err, err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error reading Archer quota: empty response")
	}
	if err != nil {
		if _, ok := err.(*quota.GetQuotasNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	archerSetQuotaResource(d, config, res.Payload)

	return nil
}

func resourceSCIEndpointQuotaV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Quota

	id := d.Id()
	req := &models.Quota{
		Endpoint: int64(d.Get("endpoint").(int)),
		Service:  int64(d.Get("service").(int)),
	}

	opts := &quota.PutQuotasProjectIDParams{
		Body:      req,
		ProjectID: id,
		Context:   ctx,
	}
	res, err := client.PutQuotasProjectID(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error updating Archer quota: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error updating Archer quota: empty response")
	}

	return resourceSCIEndpointQuotaV1Read(ctx, d, meta)
}

func resourceSCIEndpointQuotaV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Quota

	id := d.Id()
	opts := &quota.DeleteQuotasProjectIDParams{
		ProjectID: id,
		Context:   ctx,
	}
	_, err = client.DeleteQuotasProjectID(opts, c.authFunc())
	if err != nil {
		if _, ok := err.(*quota.DeleteQuotasProjectIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Archer quota: %s", err)
	}

	return nil
}

func archerSetQuotaResource(d *schema.ResourceData, config *Config, q *quota.GetQuotasProjectIDOKBody) {
	_ = d.Set("endpoint", q.Endpoint)
	_ = d.Set("service", q.Service)

	// computed
	_ = d.Set("in_use_endpoint", q.InUseEndpoint)
	_ = d.Set("in_use_service", q.InUseService)

	_ = d.Set("region", GetRegion(d, config))
}
