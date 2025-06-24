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
	"github.com/sapcc/andromeda/client/pools"
	"github.com/sapcc/andromeda/models"
)

func resourceSCIGSLBPoolV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIGSLBPoolV1Create,
		ReadContext:   resourceSCIGSLBPoolV1Read,
		UpdateContext: resourceSCIGSLBPoolV1Update,
		DeleteContext: resourceSCIGSLBPoolV1Delete,
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
			"domains": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// computed
			"members": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"monitors": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceSCIGSLBPoolV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Pools

	// Create the pool
	adminStateUp := d.Get("admin_state_up").(bool)
	pool := &models.Pool{
		AdminStateUp: &adminStateUp,
	}
	if v, ok := d.GetOk("domains"); ok {
		pool.Domains = expandToStrFmtUUIDSlice(v.([]any))
	}
	if v, ok := d.GetOk("name"); ok && v != "" {
		pool.Name = ptr(v.(string))
	}
	if v, ok := d.GetOk("project_id"); ok && v != "" {
		pool.ProjectID = ptr(v.(string))
	}

	opts := &pools.PostPoolsParams{
		Pool: pools.PostPoolsBody{
			Pool: pool,
		},
		Context: ctx,
	}
	res, err := client.PostPools(opts)
	if err != nil {
		return diag.Errorf("error creating Andromeda pool: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Pool == nil {
		return diag.Errorf("error creating Andromeda pool: empty response")
	}

	log.Printf("[DEBUG] Created Andromeda pool: %v", res)

	id := string(res.Payload.Pool.ID)
	d.SetId(id)

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := models.PoolProvisioningStatusACTIVE
	pending := models.PoolProvisioningStatusPENDINGCREATE
	pool, err = andromedaWaitForPool(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetPoolResource(d, config, pool)

	return nil
}

func resourceSCIGSLBPoolV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Pools

	id := d.Id()
	pool, err := andromedaGetPool(ctx, client, id)
	if err != nil {
		if _, ok := err.(*pools.GetPoolsPoolIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	andromedaSetPoolResource(d, config, pool)

	return nil
}

func resourceSCIGSLBPoolV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Pools

	id := d.Id()
	pool := &models.Pool{}

	if d.HasChange("admin_state_up") {
		v := d.Get("admin_state_up").(bool)
		pool.AdminStateUp = &v
	}
	if d.HasChange("domains") {
		v := d.Get("domains").([]any)
		pool.Domains = expandToStrFmtUUIDSlice(v)
	}
	if d.HasChange("name") {
		v := d.Get("name").(string)
		pool.Name = &v
	}
	if d.HasChange("project_id") {
		v := d.Get("project_id").(string)
		pool.ProjectID = &v
	}

	opts := &pools.PutPoolsPoolIDParams{
		Pool: pools.PutPoolsPoolIDBody{
			Pool: pool,
		},
		PoolID:  strfmt.UUID(id),
		Context: ctx,
	}
	_, err = client.PutPoolsPoolID(opts)
	if err != nil {
		return diag.Errorf("error updating Andromeda pool: %s", err)
	}

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutUpdate)
	target := models.PoolProvisioningStatusACTIVE
	pending := models.PoolProvisioningStatusPENDINGUPDATE
	pool, err = andromedaWaitForPool(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetPoolResource(d, config, pool)

	return nil
}

func resourceSCIGSLBPoolV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Pools

	id := d.Id()
	opts := &pools.DeletePoolsPoolIDParams{
		PoolID:  strfmt.UUID(id),
		Context: ctx,
	}
	_, err = client.DeletePoolsPoolID(opts)
	if err != nil {
		if _, ok := err.(*pools.DeletePoolsPoolIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Andromeda pool: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := "DELETED"
	pending := models.PoolProvisioningStatusPENDINGDELETE
	_, err = andromedaWaitForPool(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func andromedaWaitForPool(ctx context.Context, client pools.ClientService, id, target, pending string, timeout time.Duration) (*models.Pool, error) {
	log.Printf("[DEBUG] Waiting for %s pool to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    []string{pending},
		Refresh:    andromedaGetPoolStatus(ctx, client, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	pool, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*pools.GetPoolsPoolIDNotFound); ok && target == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s pool to become %s: %s", id, target, err)
	}

	return pool.(*models.Pool), nil
}

func andromedaGetPoolStatus(ctx context.Context, client pools.ClientService, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		pool, err := andromedaGetPool(ctx, client, id)
		if err != nil {
			return nil, "", err
		}

		return pool, pool.ProvisioningStatus, nil
	}
}

func andromedaGetPool(ctx context.Context, client pools.ClientService, id string) (*models.Pool, error) {
	opts := &pools.GetPoolsPoolIDParams{
		PoolID:  strfmt.UUID(id),
		Context: ctx,
	}
	res, err := client.GetPoolsPoolID(opts)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil || res.Payload.Pool == nil {
		return nil, fmt.Errorf("error reading Andromeda pool: empty response")
	}

	return res.Payload.Pool, nil
}

func andromedaSetPoolResource(d *schema.ResourceData, config *Config, pool *models.Pool) {
	_ = d.Set("admin_state_up", ptrValue(pool.AdminStateUp))
	_ = d.Set("domains", pool.Domains)
	_ = d.Set("name", ptrValue(pool.Name))
	_ = d.Set("project_id", ptrValue(pool.ProjectID))

	// computed
	_ = d.Set("members", pool.Members)
	_ = d.Set("monitors", pool.Monitors)
	_ = d.Set("provisioning_status", pool.ProvisioningStatus)
	_ = d.Set("status", pool.Status)
	_ = d.Set("created_at", pool.CreatedAt.String())
	_ = d.Set("updated_at", pool.UpdatedAt.String())

	_ = d.Set("region", GetRegion(d, config))
}
