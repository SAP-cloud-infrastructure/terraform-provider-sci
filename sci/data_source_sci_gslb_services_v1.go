package sci

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/andromeda/client/administrative"
	"github.com/sapcc/andromeda/models"
)

func dataSourceSCIGSLBServicesV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIGSLBServicesV1Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"heartbeat": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"metadata": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"provider": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rpc_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceSCIGSLBServicesV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Administrative

	opts := &administrative.GetServicesParams{
		Context: ctx,
	}
	res, err := client.GetServices(opts)
	if err != nil {
		return diag.Errorf("error fetching Andromeda services: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Services == nil {
		return diag.Errorf("error fetching Andromeda services: empty response")
	}

	id := andromedaServicesHash(res.Payload.Services)
	d.SetId(id)
	_ = d.Set("services", andromedaFlattenServices(res.Payload.Services))
	_ = d.Set("region", GetRegion(d, config))

	return diag.FromErr(err)
}

func andromedaFlattenServices(services []*models.Service) []map[string]any {
	res := make([]map[string]any, len(services))
	for i, service := range services {
		res[i] = map[string]any{
			"heartbeat":   service.Heartbeat.String(),
			"host":        service.Host,
			"id":          service.ID,
			"metadata":    service.Metadata,
			"provider":    service.Provider,
			"rpc_address": service.RPCAddress,
			"type":        service.Type,
			"version":     service.Version,
		}
	}
	return res
}

func andromedaServicesHash(services []*models.Service) string {
	h := sha256.New()
	for _, service := range services {
		b, _ := service.MarshalBinary()
		h.Write(b)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
