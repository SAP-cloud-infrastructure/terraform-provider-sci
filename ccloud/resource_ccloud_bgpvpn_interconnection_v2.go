package ccloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/networking/v2/bgpvpn/interconnections"
)

func resourceCCloudBGPVPNInterconnectionV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCloudBGPVPNInterconnectionV2Create,
		ReadContext:   resourceCCloudBGPVPNInterconnectionV2Read,
		UpdateContext: resourceCCloudBGPVPNInterconnectionV2Update,
		DeleteContext: resourceCCloudBGPVPNInterconnectionV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"bgpvpn",
				}, false),
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"local_resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"remote_resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"remote_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"remote_interconnection_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_parameters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_id": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"remote_parameters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_id": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceCCloudBGPVPNInterconnectionV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	networkingClient, err := config.NetworkingV2Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack networking client: %s", err)
	}

	createOpts := interconnections.CreateOpts{
		Name:                    d.Get("name").(string),
		ProjectID:               d.Get("project_id").(string),
		Type:                    d.Get("type").(string),
		LocalResourceID:         d.Get("local_resource_id").(string),
		RemoteResourceID:        d.Get("remote_resource_id").(string),
		RemoteRegion:            d.Get("remote_region").(string),
		RemoteInterconnectionID: d.Get("remote_interconnection_id").(string),
	}

	log.Printf("[DEBUG] Create BGP VPN interconnection: %#v", createOpts)

	interConn, err := interconnections.Create(networkingClient, createOpts).Extract()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] BGP VPN interconnection created: %#v", interConn)

	d.SetId(interConn.ID)

	return resourceCCloudBGPVPNInterconnectionV2Read(ctx, d, meta)
}

func resourceCCloudBGPVPNInterconnectionV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	networkingClient, err := config.NetworkingV2Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack networking client: %s", err)
	}

	interConn, err := interconnections.Get(networkingClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "interconnection"))
	}

	log.Printf("[DEBUG] Read OpenStack BPG VPN interconnection %s: %#v", d.Id(), interConn)

	d.Set("name", interConn.Name)
	d.Set("type", interConn.Type)
	d.Set("project_id", interConn.ProjectID)
	d.Set("local_resource_id", interConn.LocalResourceID)
	d.Set("remote_resource_id", interConn.RemoteResourceID)
	d.Set("remote_region", interConn.RemoteRegion)
	d.Set("remote_interconnection_id", interConn.RemoteInterconnectionID)
	d.Set("state", interConn.State)
	d.Set("local_parameters", []map[string][]string{{"project_id": interConn.LocalParameters.ProjectID}})
	d.Set("remote_parameters", []map[string][]string{{"project_id": interConn.RemoteParameters.ProjectID}})
	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceCCloudBGPVPNInterconnectionV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	networkingClient, err := config.NetworkingV2Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack networking client: %s", err)
	}

	opts := interconnections.UpdateOpts{}

	var hasChange bool

	if d.HasChange("name") {
		name := d.Get("name").(string)
		opts.Name = &name
		hasChange = true
	}

	if d.HasChange("state") {
		state := d.Get("state").(string)
		opts.State = &state
		hasChange = true
	}

	log.Printf("[DEBUG] Updating BGP VPN interconnection with id %s: %#v", d.Id(), opts)

	if hasChange {
		_, err = interconnections.Update(networkingClient, d.Id(), opts).Extract()
		if err != nil {
			return diag.FromErr(err)
		}

		log.Printf("[DEBUG] Updated BGP VPN interconnection with id %s", d.Id())
	}

	return resourceCCloudBGPVPNInterconnectionV2Read(ctx, d, meta)
}

func resourceCCloudBGPVPNInterconnectionV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Destroy interconnection: %s", d.Id())

	config := meta.(*Config)
	networkingClient, err := config.NetworkingV2Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack networking client: %s", err)
	}

	err = interconnections.Delete(networkingClient, d.Id()).Err
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
