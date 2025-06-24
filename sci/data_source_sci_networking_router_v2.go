// This augments the following upstream sources to include the SAP Cloud Infrastructure specific external_port_id
// https://github.com/terraform-provider-openstack/terraform-provider-openstack/blob/74d82f6ce503df74a5e63ac2491e837dc296a82b/openstack/data_source_openstack_networking_router_v2.go
// https://github.com/gophercloud/gophercloud/blob/39fc33cbe7c0176655a409e5fd6cbccae23bfb18/openstack/networking/v2/extensions/layer3/routers/results.go

package sci

import (
	"context"
	"log"
	"strings"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type GatewayInfo struct {
	NetworkID        string                    `json:"network_id,omitempty"`
	EnableSNAT       *bool                     `json:"enable_snat,omitempty"`
	ExternalFixedIPs []routers.ExternalFixedIP `json:"external_fixed_ips,omitempty"`
	QoSPolicyID      string                    `json:"qos_policy_id,omitempty"`
	ExternalPortID   string                    `json:"external_port_id,omitempty"`
}

type ccRouter struct {
	CCGatewayInfo GatewayInfo `json:"external_gateway_info"`
	routers.Router
}

func dataSourceSCINetworkingRouterV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCINetworkingRouterV2Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"router_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"distributed": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"external_network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_snat": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"availability_zone_hints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"external_fixed_ip": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSCINetworkingRouterV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	networkingClient, err := config.NetworkingV2Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack networking client: %s", err)
	}

	listOpts := routers.ListOpts{}

	if v, ok := d.GetOk("router_id"); ok {
		listOpts.ID = v.(string)
	}

	if v, ok := d.GetOk("name"); ok {
		listOpts.Name = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		listOpts.Description = v.(string)
	}

	if v, ok := getOkExists(d, "admin_state_up"); ok {
		asu := v.(bool)
		listOpts.AdminStateUp = &asu
	}

	if v, ok := getOkExists(d, "distributed"); ok {
		dist := v.(bool)
		listOpts.Distributed = &dist
	}

	if v, ok := d.GetOk("status"); ok {
		listOpts.Status = v.(string)
	}

	if v, ok := d.GetOk("tenant_id"); ok {
		listOpts.TenantID = v.(string)
	}

	tags := expandObjectTags(d)
	if len(tags) > 0 {
		listOpts.Tags = strings.Join(tags, ",")
	}

	pages, err := routers.List(networkingClient, listOpts).AllPages(ctx)
	if err != nil {
		return diag.Errorf("Unable to list Routers: %s", err)
	}

	var allRouters []ccRouter
	err = routers.ExtractRoutersInto(pages, &allRouters)
	if err != nil {
		return diag.Errorf("Unable to retrieve Routers: %s", err)
	}

	if len(allRouters) < 1 {
		return diag.Errorf("No Router found")
	}

	if len(allRouters) > 1 {
		return diag.Errorf("More than one Router found")
	}

	router := allRouters[0]

	log.Printf("[DEBUG] Retrieved Router %s: %+v", router.ID, router)
	d.SetId(router.ID)

	_ = d.Set("name", router.Name)
	_ = d.Set("description", router.Description)
	_ = d.Set("admin_state_up", router.AdminStateUp)
	_ = d.Set("distributed", router.Distributed)
	_ = d.Set("status", router.Status)
	_ = d.Set("tenant_id", router.TenantID)
	_ = d.Set("external_network_id", router.CCGatewayInfo.NetworkID)
	_ = d.Set("external_port_id", router.CCGatewayInfo.ExternalPortID)
	_ = d.Set("enable_snat", router.CCGatewayInfo.EnableSNAT)
	_ = d.Set("all_tags", router.Tags)
	_ = d.Set("region", GetRegion(d, config))

	if err := d.Set("availability_zone_hints", router.AvailabilityZoneHints); err != nil {
		log.Printf("[DEBUG] Unable to set availability_zone_hints: %s", err)
	}

	externalFixedIPs := make([]map[string]string, 0, len(router.GatewayInfo.ExternalFixedIPs))
	for _, v := range router.GatewayInfo.ExternalFixedIPs {
		externalFixedIPs = append(externalFixedIPs, map[string]string{
			"subnet_id":  v.SubnetID,
			"ip_address": v.IPAddress,
		})
	}
	if err = d.Set("external_fixed_ip", externalFixedIPs); err != nil {
		log.Printf("[DEBUG] Unable to set external_fixed_ip: %s", err)
	}
	return nil
}
