package ccloud

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
	"github.com/sapcc/andromeda/client/monitors"
	"github.com/sapcc/andromeda/models"
)

func resourceCCloudGSLBMonitorV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCloudGSLBMonitorV1Create,
		ReadContext:   resourceCCloudGSLBMonitorV1Read,
		UpdateContext: resourceCCloudGSLBMonitorV1Update,
		DeleteContext: resourceCCloudGSLBMonitorV1Delete,
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
			"interval": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"receive": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"send": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"type": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"ICMP", "HTTP", "HTTPS", "TCP", "UDP",
				}, false),
				Optional: true,
				Default:  "ICMP",
			},

			// computed
			"provisioning_status": {
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

func resourceCCloudGSLBMonitorV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Monitors

	// Create the member
	adminStateUp := d.Get("admin_state_up").(bool)
	monitor := &models.Monitor{
		AdminStateUp: &adminStateUp,
	}
	if v, ok := d.GetOk("interval"); ok && v != 0 {
		monitor.Interval = ptr(int64(v.(int)))
	}
	if v, ok := d.GetOk("name"); ok && v != "" {
		monitor.Name = ptr(v.(string))
	}
	if v, ok := d.GetOk("pool_id"); ok && v != "" {
		v := strfmt.UUID(v.(string))
		monitor.PoolID = &v
	}
	if v, ok := d.GetOk("project_id"); ok && v != "" {
		monitor.ProjectID = ptr(v.(string))
	}
	if v, ok := d.GetOk("receive"); ok && v != "" {
		monitor.Receive = ptr(v.(string))
	}
	if v, ok := d.GetOk("send"); ok && v != "" {
		monitor.Send = ptr(v.(string))
	}
	if v, ok := d.GetOk("timeout"); ok && v != 0 {
		monitor.Timeout = ptr(int64(v.(int)))
	}
	if v, ok := d.GetOk("type"); ok && v != "" {
		monitor.Type = ptr(v.(string))
	}

	opts := &monitors.PostMonitorsParams{
		Monitor: monitors.PostMonitorsBody{
			Monitor: monitor,
		},
		Context: ctx,
	}
	res, err := client.PostMonitors(opts)
	if err != nil {
		return diag.Errorf("error creating Andromeda monitor: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Monitor == nil {
		return diag.Errorf("error creating Andromeda monitor: empty response")
	}

	log.Printf("[DEBUG] Created Andromeda monitor: %s", res)

	id := string(res.Payload.Monitor.ID)
	d.SetId(id)

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := models.MonitorProvisioningStatusACTIVE
	pending := models.MonitorProvisioningStatusPENDINGCREATE
	monitor, err = andromedaWaitForMonitor(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetMonitorResource(d, config, monitor)

	return nil
}

func resourceCCloudGSLBMonitorV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Monitors

	id := d.Id()
	monitor, err := andromedaGetMonitor(ctx, client, id)
	if err != nil {
		if _, ok := err.(*monitors.GetMonitorsMonitorIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	andromedaSetMonitorResource(d, config, monitor)

	return nil
}

func resourceCCloudGSLBMonitorV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Monitors

	id := d.Id()
	monitor := &models.Monitor{}

	if d.HasChange("admin_state_up") {
		v := d.Get("admin_state_up").(bool)
		monitor.AdminStateUp = &v
	}
	if d.HasChange("interval") {
		v := int64(d.Get("interval").(int))
		monitor.Interval = &v
	}
	if d.HasChange("name") {
		v := d.Get("name").(string)
		monitor.Name = &v
	}
	if d.HasChange("pool_id") {
		v := strfmt.UUID(d.Get("pool_id").(string))
		monitor.PoolID = &v
	}
	if d.HasChange("receive") {
		v := d.Get("receive").(string)
		monitor.Receive = &v
	}
	if d.HasChange("send") {
		v := d.Get("send").(string)
		monitor.Send = &v
	}
	if d.HasChange("timeout") {
		v := int64(d.Get("timeout").(int))
		monitor.Timeout = &v
	}
	if d.HasChange("type") {
		v := d.Get("type").(string)
		monitor.Type = &v
	}

	opts := &monitors.PutMonitorsMonitorIDParams{
		Monitor: monitors.PutMonitorsMonitorIDBody{
			Monitor: monitor,
		},
		MonitorID: strfmt.UUID(id),
		Context:   ctx,
	}
	_, err = client.PutMonitorsMonitorID(opts)
	if err != nil {
		return diag.Errorf("error updating Andromeda monitor: %s", err)
	}

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutUpdate)
	target := models.MonitorProvisioningStatusACTIVE
	pending := models.MonitorProvisioningStatusPENDINGUPDATE
	_, err = andromedaWaitForMonitor(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetMonitorResource(d, config, monitor)

	return nil
}

func resourceCCloudGSLBMonitorV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Monitors

	id := d.Id()
	opts := &monitors.DeleteMonitorsMonitorIDParams{
		MonitorID: strfmt.UUID(id),
		Context:   ctx,
	}
	_, err = client.DeleteMonitorsMonitorID(opts)
	if err != nil {
		if _, ok := err.(*monitors.DeleteMonitorsMonitorIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Andromeda monitor: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := "DELETED"
	pending := models.MonitorProvisioningStatusPENDINGDELETE
	_, err = andromedaWaitForMonitor(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func andromedaWaitForMonitor(ctx context.Context, client monitors.ClientService, id, target, pending string, timeout time.Duration) (*models.Monitor, error) {
	log.Printf("[DEBUG] Waiting for %s monitor to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    []string{pending},
		Refresh:    andromedaGetMonitorStatus(ctx, client, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	monitor, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*monitors.GetMonitorsMonitorIDNotFound); ok && target == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s monitor to become %s: %s", id, target, err)
	}

	return monitor.(*models.Monitor), nil
}

func andromedaGetMonitorStatus(ctx context.Context, client monitors.ClientService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		monitor, err := andromedaGetMonitor(ctx, client, id)
		if err != nil {
			return nil, "", err
		}

		return monitor, monitor.ProvisioningStatus, nil
	}
}

func andromedaGetMonitor(ctx context.Context, client monitors.ClientService, id string) (*models.Monitor, error) {
	opts := &monitors.GetMonitorsMonitorIDParams{
		MonitorID: strfmt.UUID(id),
		Context:   ctx,
	}
	res, err := client.GetMonitorsMonitorID(opts)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil || res.Payload.Monitor == nil {
		return nil, fmt.Errorf("error reading Andromeda monitor: empty response")
	}

	return res.Payload.Monitor, nil
}

func andromedaSetMonitorResource(d *schema.ResourceData, config *Config, monitor *models.Monitor) {
	d.Set("admin_state_up", ptrValue(monitor.AdminStateUp))
	d.Set("interval", ptrValue(monitor.Interval))
	d.Set("name", ptrValue(monitor.Name))
	d.Set("pool_id", ptrValue(monitor.PoolID))
	d.Set("project_id", ptrValue(monitor.ProjectID))
	d.Set("receive", ptrValue(monitor.Receive))
	d.Set("send", ptrValue(monitor.Send))
	d.Set("timeout", ptrValue(monitor.Timeout))
	d.Set("type", ptrValue(monitor.Type))

	// computed
	d.Set("provisioning_status", monitor.ProvisioningStatus)
	d.Set("created_at", monitor.CreatedAt.String())
	d.Set("updated_at", monitor.UpdatedAt.String())

	d.Set("region", GetRegion(d, config))
}
