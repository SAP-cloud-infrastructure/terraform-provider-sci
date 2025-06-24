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
	"github.com/sapcc/archer/client/service"
	"github.com/sapcc/archer/models"
)

func resourceSCIEndpointAcceptV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIEndpointAcceptV1Create,
		ReadContext:   resourceSCIEndpointAcceptV1Read,
		DeleteContext: resourceSCIEndpointAcceptV1Delete,
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

			// computed
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSCIEndpointAcceptV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Service

	// Accept the service endpoint consumer
	serviceID := d.Get("service_id").(string)
	endpointID := d.Get("endpoint_id").(string)
	req := &models.EndpointConsumerList{
		EndpointIds: []strfmt.UUID{strfmt.UUID(endpointID)},
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

	id := fmt.Sprintf("%s/%s", serviceID, endpointID)
	d.SetId(id)

	// waiting for AVAILABLE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := []string{
		string(models.EndpointStatusAVAILABLE),
	}
	pending := []string{
		string(models.EndpointStatusPENDINGCREATE),
		string(models.EndpointStatusPENDINGAPPROVAL),
	}
	ec, err := archerWaitForServiceEndpointConsumer(ctx, c, endpointID, serviceID, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	archerSetServiceEndpointConsumer(d, config, endpointID, ec)

	return nil
}

func resourceSCIEndpointAcceptV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}

	serviceID, id, err := parsePairedIDs(d.Id(), "sci_endpoint_accept_v1")
	if err != nil {
		return diag.FromErr(err)
	}

	ec, err := archerGetServiceEndpointConsumer(ctx, c, id, serviceID)
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

func resourceSCIEndpointAcceptV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Service

	serviceID, id, err := parsePairedIDs(d.Id(), "sci_endpoint_accept_v1")
	if err != nil {
		return diag.FromErr(err)
	}

	req := &models.EndpointConsumerList{
		EndpointIds: []strfmt.UUID{strfmt.UUID(id)},
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
	target := []string{
		"DELETED",
		string(models.EndpointStatusREJECTED),
	}
	pending := []string{
		string(models.EndpointStatusPENDINGREJECTED),
		string(models.EndpointStatusPENDINGDELETE),
	}
	_, err = archerWaitForServiceEndpointConsumer(ctx, c, id, serviceID, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func archerWaitForServiceEndpointConsumer(ctx context.Context, c *archer, id, serviceID string, target, pending []string, timeout time.Duration) (*models.EndpointConsumer, error) {
	log.Printf("[DEBUG] Waiting for %s endpoint to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     target,
		Pending:    pending,
		Refresh:    archerGetServiceEndpointConsumerStatus(ctx, c, id, serviceID),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	ec, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*service.GetServiceServiceIDEndpointsNotFound); ok && sliceContains(target, "DELETED") {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s endpoint to become %s: %s", id, target, err)
	}

	return ec.(*models.EndpointConsumer), nil
}

func archerGetServiceEndpointConsumerStatus(ctx context.Context, c *archer, id, serviceID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		ec, err := archerGetServiceEndpointConsumer(ctx, c, id, serviceID)
		if err != nil {
			return nil, "", err
		}

		return ec, string(ec.Status), nil
	}
}

func archerGetServiceEndpointConsumer(ctx context.Context, c *archer, id, serviceID string) (*models.EndpointConsumer, error) {
	opts := &service.GetServiceServiceIDEndpointsParams{
		ServiceID: strfmt.UUID(serviceID),
		Context:   ctx,
	}
	res, err := c.Service.GetServiceServiceIDEndpoints(opts, c.authFunc())
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil {
		return nil, fmt.Errorf("error reading Archer endpoint: empty response")
	}

	for _, v := range res.Payload.Items {
		if v.ID == strfmt.UUID(id) {
			return v, nil
		}
	}

	return nil, &service.GetServiceServiceIDEndpointsNotFound{}
}

func archerSetServiceEndpointConsumer(d *schema.ResourceData, config *Config, id string, consumer *models.EndpointConsumer) {
	_ = d.Set("endpoint_id", id)
	_ = d.Set("status", consumer.Status)
	_ = d.Set("region", GetRegion(d, config))
}
