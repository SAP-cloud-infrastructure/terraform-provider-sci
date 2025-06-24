package sci

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/archer/client/service"
	"github.com/sapcc/archer/models"
)

func dataSourceSCIEndpointServiceV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIEndpointServiceV1Read,

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
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip_addresses": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPAddress,
				},
				Optional: true,
			},
			"all_ip_addresses": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPAddress,
				},
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"service_provider": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"tenant", "cp",
				}, false),
				Optional: true,
				Computed: true,
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
			},
			"all_tags": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"host": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
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

func dataSourceSCIEndpointServiceV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Service

	// List the services
	listOpts := &service.GetServiceParams{
		Tags: expandToStringSlice(d.Get("tags").([]any)),
	}
	if v, ok := d.GetOk("project_id"); ok {
		v := v.(string)
		listOpts.ProjectID = &v
	}

	services, err := client.GetService(listOpts, c.authFunc())
	if err != nil {
		return diag.Errorf("error listing Archer services: %s", err)
	}

	if services.Payload == nil || len(services.Payload.Items) == 0 {
		return diag.Errorf("Archer services not found")
	}

	filteredServices := make([]models.Service, 0, len(services.Payload.Items))

	// define filter values
	var name, description, availabilityZone, networkID, provider, visibility, host, status *string
	var enabled, proxyProtocol, requireApproval *bool
	var port *int32
	var ipAddresses []string

	if v, ok := d.GetOk("name"); ok {
		name = ptr(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		description = ptr(v.(string))
	}
	if v, ok := d.GetOk("availability_zone"); ok {
		availabilityZone = ptr(v.(string))
	}
	if v, ok := d.GetOk("network_id"); ok {
		networkID = ptr(v.(string))
	}
	if v, ok := d.GetOk("service_provider"); ok {
		provider = ptr(v.(string))
	}
	if v, ok := d.GetOk("visibility"); ok {
		visibility = ptr(v.(string))
	}
	if v, ok := d.GetOk("host"); ok {
		host = ptr(v.(string))
	}
	if v, ok := d.GetOk("status"); ok {
		status = ptr(v.(string))
	}

	if v, ok := d.GetOk("enabled"); ok {
		enabled = ptr(v.(bool))
	}
	if v, ok := d.GetOk("proxy_protocol"); ok {
		proxyProtocol = ptr(v.(bool))
	}
	if v, ok := d.GetOk("require_approval"); ok {
		requireApproval = ptr(v.(bool))
	}

	if v, ok := d.GetOk("port"); ok {
		port = ptr(int32(v.(int)))
	}

	if v, ok := d.GetOk("ip_addresses"); ok {
		ipAddresses = expandToStringSlice(v.([]any))
	}

ItemsLoop:
	for _, svc := range services.Payload.Items {
		if svc == nil {
			continue
		}
		if name != nil && *name != svc.Name {
			continue
		}
		if description != nil && *description != svc.Description {
			continue
		}
		if availabilityZone != nil && *availabilityZone != ptrValue(svc.AvailabilityZone) {
			continue
		}
		if networkID != nil && *networkID != string(ptrValue(svc.NetworkID)) {
			continue
		}
		if provider != nil && *provider != ptrValue(svc.Provider) {
			continue
		}
		if visibility != nil && *visibility != ptrValue(svc.Visibility) {
			continue
		}
		if enabled != nil && *enabled != ptrValue(svc.Enabled) {
			continue
		}
		if proxyProtocol != nil && *proxyProtocol != ptrValue(svc.ProxyProtocol) {
			continue
		}
		if requireApproval != nil && *requireApproval != ptrValue(svc.RequireApproval) {
			continue
		}
		if port != nil && *port != svc.Port {
			continue
		}
		if host != nil && *host != ptrValue(svc.Host) {
			continue
		}
		if status != nil && *status != svc.Status {
			continue
		}
		svcIPAddresses := flattenToStrFmtIPv4Slice(svc.IPAddresses)
		for _, ip := range ipAddresses {
			if !sliceContains(svcIPAddresses, ip) {
				continue ItemsLoop
			}
		}
		filteredServices = append(filteredServices, *svc)
	}

	if len(filteredServices) == 0 {
		return diag.Errorf("Archer services not found")
	}

	if len(filteredServices) > 1 {
		return diag.Errorf("found more than one Archer services: %v", filteredServices)
	}

	svc := services.Payload.Items[0]

	d.SetId(string(svc.ID))

	_ = d.Set("enabled", ptrValue(svc.Enabled))
	_ = d.Set("all_ip_addresses", flattenToStrFmtIPv4Slice(svc.IPAddresses))
	_ = d.Set("name", svc.Name)
	_ = d.Set("description", svc.Description)
	_ = d.Set("port", svc.Port)
	_ = d.Set("network_id", ptrValue(svc.NetworkID))
	_ = d.Set("project_id", svc.ProjectID)
	_ = d.Set("all_tags", svc.Tags)
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

	return nil
}
