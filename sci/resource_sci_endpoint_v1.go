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
	"github.com/sapcc/archer/client/endpoint"
	"github.com/sapcc/archer/models"
)

func resourceSCIEndpointV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIEndpointV1Create,
		ReadContext:   resourceSCIEndpointV1Read,
		UpdateContext: resourceSCIEndpointV1Update,
		DeleteContext: resourceSCIEndpointV1Delete,
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
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"target.0.port", "target.0.subnet"},
						},
						"port": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"target.0.network", "target.0.subnet"},
						},
						"subnet": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"target.0.network", "target.0.port"},
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			// computed
			"ip_address": {
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

func resourceSCIEndpointV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Endpoint

	// Create the endpoint
	ept := &models.Endpoint{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ProjectID:   models.Project(d.Get("project_id").(string)),
		ServiceID:   strfmt.UUID(d.Get("service_id").(string)),
		Tags:        expandToStringSlice(d.Get("tags").([]any)),
		Target:      flattenEndpointTarget(d.Get("target").([]any)),
	}

	opts := &endpoint.PostEndpointParams{
		Body:    ept,
		Context: ctx,
	}
	res, err := client.PostEndpoint(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error creating Archer endpoint: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error creating Archer endpoint: empty response")
	}

	log.Printf("[DEBUG] Created Archer endpoint: %v", res)

	id := string(res.Payload.ID)
	d.SetId(id)

	// waiting for AVAILABLE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := []string{
		string(models.EndpointStatusAVAILABLE),
		string(models.EndpointStatusPENDINGAPPROVAL),
	}
	pending := string(models.EndpointStatusPENDINGCREATE)
	ept, err = archerWaitForEndpoint(ctx, c, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	archerSetEndpointResource(d, config, ept)

	return nil
}

func resourceSCIEndpointV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}

	id := d.Id()
	ept, err := archerGetEndpoint(ctx, c, id)
	if err != nil {
		if _, ok := err.(*endpoint.GetEndpointEndpointIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	archerSetEndpointResource(d, config, ept)

	return nil
}

func resourceSCIEndpointV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Endpoint

	id := d.Id()
	ept := endpoint.PutEndpointEndpointIDBody{}

	if d.HasChange("name") {
		ept.Name = ptr(d.Get("name").(string))
	}
	if d.HasChange("description") {
		ept.Description = ptr(d.Get("description").(string))
	}
	if d.HasChange("tags") {
		ept.Tags = expandToStringSlice(d.Get("tags").([]any))
	}

	opts := &endpoint.PutEndpointEndpointIDParams{
		Body:       ept,
		EndpointID: strfmt.UUID(id),
		Context:    ctx,
	}
	res, err := client.PutEndpointEndpointID(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error updating Archer endpoint: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error updating Archer endpoint: empty response")
	}

	archerSetEndpointResource(d, config, res.Payload)

	return nil
}

func resourceSCIEndpointV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Endpoint

	id := d.Id()
	opts := &endpoint.DeleteEndpointEndpointIDParams{
		EndpointID: strfmt.UUID(id),
		Context:    ctx,
	}
	_, err = client.DeleteEndpointEndpointID(opts, c.authFunc())
	if err != nil {
		if _, ok := err.(*endpoint.DeleteEndpointEndpointIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Archer endpoint: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := []string{"DELETED"}
	pending := string(models.EndpointStatusPENDINGDELETE)
	_, err = archerWaitForEndpoint(ctx, c, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func archerWaitForEndpoint(ctx context.Context, c *archer, id string, target []string, pending string, timeout time.Duration) (*models.Endpoint, error) {
	log.Printf("[DEBUG] Waiting for %s endpoint to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     target,
		Pending:    []string{pending},
		Refresh:    archerGetEndpointStatus(ctx, c, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	ept, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*endpoint.GetEndpointEndpointIDNotFound); ok && target[0] == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s endpoint to become %s: %s", id, target, err)
	}

	return ept.(*models.Endpoint), nil
}

func archerGetEndpointStatus(ctx context.Context, c *archer, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		endpoint, err := archerGetEndpoint(ctx, c, id)
		if err != nil {
			return nil, "", err
		}

		return endpoint, string(endpoint.Status), nil
	}
}

func archerGetEndpoint(ctx context.Context, c *archer, id string) (*models.Endpoint, error) {
	opts := &endpoint.GetEndpointEndpointIDParams{
		EndpointID: strfmt.UUID(id),
		Context:    ctx,
	}
	res, err := c.Endpoint.GetEndpointEndpointID(opts, c.authFunc())
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil {
		return nil, fmt.Errorf("error reading Archer endpoint: empty response")
	}

	return res.Payload, nil
}

func archerSetEndpointResource(d *schema.ResourceData, config *Config, ept *models.Endpoint) {
	_ = d.Set("name", ept.Name)
	_ = d.Set("description", ept.Description)
	_ = d.Set("service_id", ept.ServiceID)
	_ = d.Set("project_id", ept.ProjectID)
	_ = d.Set("ip_address", ept.IPAddress)
	_ = d.Set("tags", ept.Tags)
	_ = d.Set("target", expandEndpointTarget(ept.Target))

	// computed
	_ = d.Set("status", ept.Status)
	_ = d.Set("created_at", ept.CreatedAt.String())
	_ = d.Set("updated_at", ept.UpdatedAt.String())

	_ = d.Set("region", GetRegion(d, config))
}

func expandEndpointTarget(target models.EndpointTarget) []map[string]string {
	return []map[string]string{
		{
			"network": string(ptrValue(target.Network)),
			"port":    string(ptrValue(target.Port)),
			"subnet":  string(ptrValue(target.Subnet)),
		},
	}
}

func flattenEndpointTarget(targets []any) models.EndpointTarget {
	res := models.EndpointTarget{}
	for _, t := range targets {
		m := t.(map[string]any)
		if v, ok := m["network"].(string); ok && v != "" {
			res.Network = ptr(strfmt.UUID(v))
		}
		if v, ok := m["port"].(string); ok && v != "" {
			res.Port = ptr(strfmt.UUID(v))
		}
		if v, ok := m["subnet"].(string); ok && v != "" {
			res.Subnet = ptr(strfmt.UUID(v))
		}
	}
	return res
}
