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
	"github.com/sapcc/archer/client/service"
	"github.com/sapcc/archer/models"
)

func resourceSCIEndpointServiceV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIEndpointServiceV1Create,
		ReadContext:   resourceSCIEndpointServiceV1Read,
		UpdateContext: resourceSCIEndpointServiceV1Update,
		DeleteContext: resourceSCIEndpointServiceV1Delete,
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
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ip_addresses": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPAddress,
				},
				Required: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
					"tenant", "cp",
				}, false),
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"proxy_protocol": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"require_approval": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"visibility": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"private", "public",
				}, false),
				Optional: true,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},

			// computed
			"host": {
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

func resourceSCIEndpointServiceV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Service

	// Create the service
	enabled := d.Get("enabled").(bool)
	networkID := strfmt.UUID(d.Get("network_id").(string))
	svc := &models.Service{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ProjectID:   models.Project(d.Get("project_id").(string)),
		Enabled:     &enabled,
		NetworkID:   &networkID,
		Port:        int32(d.Get("port").(int)),
		IPAddresses: expandToStrFmtIPv4Slice(d.Get("ip_addresses").([]any)),
	}
	if v, ok := getOkExists(d, "proxy_protocol"); ok {
		svc.ProxyProtocol = ptr(v.(bool))
	}
	if v, ok := getOkExists(d, "require_approval"); ok {
		svc.RequireApproval = ptr(v.(bool))
	}
	if v, ok := d.GetOk("availability_zone"); ok && v != "" {
		svc.AvailabilityZone = ptr(v.(string))
	}
	if v, ok := d.GetOk("svc_provider"); ok && v != "" {
		svc.Provider = ptr(v.(string))
	}
	if v, ok := d.GetOk("visibility"); ok && v != "" {
		svc.Visibility = ptr(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		svc.Tags = expandToStringSlice(v.([]any))
	}

	opts := &service.PostServiceParams{
		Body:    svc,
		Context: ctx,
	}
	res, err := client.PostService(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error creating Archer service: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error creating Archer service: empty response")
	}

	log.Printf("[DEBUG] Created Archer service: %v", res)

	id := string(res.Payload.ID)
	d.SetId(id)

	// waiting for AVAILABLE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := models.ServiceStatusAVAILABLE
	pending := models.ServiceStatusPENDINGCREATE
	svc, err = archerWaitForService(ctx, c, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	archerSetServiceResource(d, config, svc)

	return nil
}

func resourceSCIEndpointServiceV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}

	id := d.Id()
	svc, err := archerGetService(ctx, c, id)
	if err != nil {
		if _, ok := err.(*service.GetServiceServiceIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	archerSetServiceResource(d, config, svc)

	return nil
}

func resourceSCIEndpointServiceV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Service

	id := d.Id()
	svc := &models.ServiceUpdatable{
		IPAddresses: expandToStrFmtIPv4Slice(d.Get("ip_addresses").([]any)),
		Tags:        expandToStringSlice(d.Get("tags").([]any)),
	}

	if d.HasChange("enabled") {
		v := d.Get("enabled").(bool)
		svc.Enabled = &v
	}
	if d.HasChange("name") {
		v := d.Get("name").(string)
		svc.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		svc.Description = &v
	}
	if d.HasChange("port") {
		v := int32(d.Get("port").(int))
		svc.Port = &v
	}
	if d.HasChange("proxy_protocol") {
		v := d.Get("proxy_protocol").(bool)
		svc.ProxyProtocol = &v
	}
	if d.HasChange("require_approval") {
		v := d.Get("require_approval").(bool)
		svc.RequireApproval = &v
	}
	if d.HasChange("visibility") {
		v := d.Get("visibility").(string)
		svc.Visibility = &v
	}

	opts := &service.PutServiceServiceIDParams{
		Body:      svc,
		ServiceID: strfmt.UUID(id),
		Context:   ctx,
	}
	_, err = client.PutServiceServiceID(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error updating Archer service: %s", err)
	}

	// waiting for AVAILABLE status
	timeout := d.Timeout(schema.TimeoutUpdate)
	target := models.ServiceStatusAVAILABLE
	pending := models.ServiceStatusPENDINGUPDATE
	res, err := archerWaitForService(ctx, c, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	archerSetServiceResource(d, config, res)

	return nil
}

func resourceSCIEndpointServiceV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Service

	id := d.Id()
	opts := &service.DeleteServiceServiceIDParams{
		ServiceID: strfmt.UUID(id),
		Context:   ctx,
	}
	_, err = client.DeleteServiceServiceID(opts, c.authFunc())
	if err != nil {
		if _, ok := err.(*service.DeleteServiceServiceIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Archer service: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := "DELETED"
	pending := models.ServiceStatusPENDINGDELETE
	_, err = archerWaitForService(ctx, c, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func archerWaitForService(ctx context.Context, c *archer, id, target, pending string, timeout time.Duration) (*models.Service, error) {
	log.Printf("[DEBUG] Waiting for %s service to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    []string{pending},
		Refresh:    archerGetServiceStatus(ctx, c, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	svc, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*service.GetServiceServiceIDNotFound); ok && target == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s service to become %s: %s", id, target, err)
	}

	return svc.(*models.Service), nil
}

func archerGetServiceStatus(ctx context.Context, c *archer, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		service, err := archerGetService(ctx, c, id)
		if err != nil {
			return nil, "", err
		}

		return service, service.Status, nil
	}
}

func archerGetService(ctx context.Context, c *archer, id string) (*models.Service, error) {
	opts := &service.GetServiceServiceIDParams{
		ServiceID: strfmt.UUID(id),
		Context:   ctx,
	}
	res, err := c.Service.GetServiceServiceID(opts, c.authFunc())
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil {
		return nil, fmt.Errorf("error reading Archer service: empty response")
	}

	return res.Payload, nil
}

func archerSetServiceResource(d *schema.ResourceData, config *Config, svc *models.Service) {
	_ = d.Set("enabled", ptrValue(svc.Enabled))
	_ = d.Set("ip_addresses", flattenToStrFmtIPv4Slice(svc.IPAddresses))
	_ = d.Set("name", svc.Name)
	_ = d.Set("description", svc.Description)
	_ = d.Set("port", svc.Port)
	_ = d.Set("network_id", ptrValue(svc.NetworkID))
	_ = d.Set("project_id", svc.ProjectID)
	_ = d.Set("tags", svc.Tags)
	_ = d.Set("service_provider", ptrValue(svc.Provider))
	_ = d.Set("proxy_protocol", ptrValue(svc.ProxyProtocol))
	_ = d.Set("require_approval", ptrValue(svc.RequireApproval))
	_ = d.Set("visibility", ptrValue(svc.Visibility))

	// computed
	_ = d.Set("host", ptrValue(svc.Host))
	_ = d.Set("status", svc.Status)
	_ = d.Set("created_at", svc.CreatedAt.String())
	_ = d.Set("updated_at", svc.UpdatedAt.String())

	_ = d.Set("region", GetRegion(d, config))
}
