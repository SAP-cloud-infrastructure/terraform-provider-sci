package ccloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/archer/client/service"
	"github.com/sapcc/archer/models"
)

func resourceCCloudEndpointAcceptV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCloudEndpointAcceptV1Create,
		ReadContext:   resourceCCloudEndpointAcceptV1Read,
		DeleteContext: resourceCCloudEndpointAcceptV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: archerSetServiceEndpointConsumerImport,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// computed
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCCloudEndpointAcceptV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Service

	// Accept the service endpoint consumer
	serviceID := d.Get("service_id").(string)
	endpointID := d.Get("endpoint_id").(string)
	projectID := d.Get("project_id").(string)
	req := &models.EndpointConsumerList{
		EndpointIds: []strfmt.UUID{strfmt.UUID(endpointID)},
		ProjectIds:  []models.Project{models.Project(projectID)},
	}

	opts := &service.PutServiceServiceIDAcceptEndpointsParams{
		Body:      req,
		ServiceID: strfmt.UUID(serviceID),
		Context:   ctx,
	}
	res, err := client.PutServiceServiceIDAcceptEndpoints(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error accepting Archer endpoint: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error accepting Archer endpoint: empty response")
	}

	log.Printf("[DEBUG] Accepted Archer endpoint: %v", res)

	d.SetId(endpointID)

	// waiting for AVAILABLE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := string(models.EndpointStatusAVAILABLE)
	pending := []string{
		string(models.EndpointStatusPENDINGCREATE),
		string(models.EndpointStatusPENDINGAPPROVAL),
	}
	ec, err := archerWaitForServiceEndpointConsumer(ctx, c, endpointID, serviceID, projectID, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	archerSetServiceEndpointConsumer(d, config, endpointID, ec)

	return nil
}

func resourceCCloudEndpointAcceptV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}

	id := d.Id()
	serviceID := d.Get("service_id").(string)
	projectID := d.Get("project_id").(string)
	ec, err := archerGetServiceEndpointConsumer(ctx, c, id, serviceID, projectID)
	if err != nil {
		if _, ok := err.(*service.GetServiceServiceIDEndpointsNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading Archer endpoint consumer: %s", err)
	}

	archerSetServiceEndpointConsumer(d, config, id, ec)

	return nil
}

func resourceCCloudEndpointAcceptV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Service

	serviceID := d.Get("service_id").(string)
	projectID := d.Get("project_id").(string)
	req := &models.EndpointConsumerList{
		EndpointIds: []strfmt.UUID{strfmt.UUID(d.Id())},
		ProjectIds:  []models.Project{models.Project(projectID)},
	}
	opts := &service.PutServiceServiceIDRejectEndpointsParams{
		Body:      req,
		ServiceID: strfmt.UUID(serviceID),
		Context:   ctx,
	}
	_, err = client.PutServiceServiceIDRejectEndpoints(opts, c.authFunc())
	if err != nil {
		if _, ok := err.(*service.PutServiceServiceIDRejectEndpointsNotFound); ok {
			return nil
		}
		return diag.Errorf("error rejecting Archer endpoint: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := "DELETED"
	pending := []string{
		string(models.EndpointStatusPENDINGREJECTED),
		string(models.EndpointStatusPENDINGDELETE),
	}
	_, err = archerWaitForServiceEndpointConsumer(ctx, c, d.Id(), serviceID, projectID, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func archerWaitForServiceEndpointConsumer(ctx context.Context, c *archer, id, serviceID, projectID, target string, pending []string, timeout time.Duration) (*models.EndpointConsumer, error) {
	log.Printf("[DEBUG] Waiting for %s endpoint to become %s.", id, target)

	stateConf := &resource.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    archerGetServiceEndpointConsumerStatus(ctx, c, id, serviceID, projectID),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	ec, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*service.GetServiceServiceIDEndpointsNotFound); ok && target == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s endpoint to become %s: %s", id, target, err)
	}

	return ec.(*models.EndpointConsumer), nil
}

func archerGetServiceEndpointConsumerStatus(ctx context.Context, c *archer, id, serviceID, projectID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ec, err := archerGetServiceEndpointConsumer(ctx, c, id, serviceID, projectID)
		if err != nil {
			return nil, "", err
		}

		return ec, string(ec.Status), nil
	}
}

func archerGetServiceEndpointConsumer(ctx context.Context, c *archer, id, serviceID, projectID string) (*models.EndpointConsumer, error) {
	opts := &service.GetServiceServiceIDEndpointsParams{
		ServiceID: strfmt.UUID(serviceID),
	}
	res, err := c.Service.GetServiceServiceIDEndpoints(opts, c.authFunc())
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil {
		return nil, fmt.Errorf("error reading Archer endpoint: empty response")
	}

	for _, v := range res.Payload.Items {
		if v.ID == strfmt.UUID(id) && v.ProjectID == models.Project(projectID) {
			return v, nil
		}
	}

	return nil, &service.GetServiceServiceIDEndpointsNotFound{}
}

func archerSetServiceEndpointConsumer(d *schema.ResourceData, config *Config, id string, consumer *models.EndpointConsumer) {
	d.Set("endpoint_id", id)
	d.Set("project_id", consumer.ProjectID)
	d.Set("status", consumer.Status)
	d.Set("region", GetRegion(d, config))
}

func archerSetServiceEndpointConsumerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 3)
	if len(parts) != 3 {
		err := fmt.Errorf("Invalid format specified for endpoint consumer. Format must be <endpoint id>/<service id>/<project id>")
		return nil, err
	}

	id := parts[0]
	serviceID := parts[1]
	projectID := parts[2]

	d.SetId(id)
	d.Set("service_id", serviceID)
	d.Set("project_id", projectID)

	return []*schema.ResourceData{d}, nil
}
