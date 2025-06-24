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
	"github.com/sapcc/andromeda/client/datacenters"
	"github.com/sapcc/andromeda/models"
)

func resourceSCIGSLBDatacenterV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIGSLBDatacenterV1Create,
		ReadContext:   resourceSCIGSLBDatacenterV1Read,
		UpdateContext: resourceSCIGSLBDatacenterV1Update,
		DeleteContext: resourceSCIGSLBDatacenterV1Delete,
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
			"city": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"continent": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"country": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"latitude": {
				Type:     schema.TypeFloat,
				Optional: true,
				Computed: true,
			},
			"longitude": {
				Type:     schema.TypeFloat,
				Optional: true,
				Computed: true,
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
			"service_provider": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"akamai", "f5",
				}, false),
				Optional: true,
				Default:  "akamai",
			},
			"scope": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"private", "shared",
				}, false),
				Optional: true,
				Default:  "private",
			},
			"state_or_province": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// computed
			"provisioning_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"meta": {
				Type:     schema.TypeInt,
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

func resourceSCIGSLBDatacenterV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Datacenters

	// Create the datacenter
	adminStateUp := d.Get("admin_state_up").(bool)
	provider := d.Get("service_provider").(string)
	scope := d.Get("scope").(string)
	datacenter := &models.Datacenter{
		AdminStateUp: &adminStateUp,
		Provider:     provider,
		Scope:        &scope,
	}
	if v, ok := d.GetOk("city"); ok && v != "" {
		datacenter.City = ptr(v.(string))
	}
	if v, ok := d.GetOk("continent"); ok && v != "" {
		datacenter.Continent = ptr(v.(string))
	}
	if v, ok := d.GetOk("country"); ok && v != "" {
		datacenter.Country = ptr(v.(string))
	}
	if v, ok := d.GetOk("latitude"); ok && v != 0 {
		datacenter.Latitude = ptr(v.(float64))
	}
	if v, ok := d.GetOk("longitude"); ok && v != 0 {
		datacenter.Longitude = ptr(v.(float64))
	}
	if v, ok := d.GetOk("name"); ok && v != "" {
		datacenter.Name = ptr(v.(string))
	}
	if v, ok := d.GetOk("state_or_province"); ok && v != "" {
		datacenter.StateOrProvince = ptr(v.(string))
	}

	opts := &datacenters.PostDatacentersParams{
		Datacenter: datacenters.PostDatacentersBody{
			Datacenter: datacenter,
		},
		Context: ctx,
	}
	res, err := client.PostDatacenters(opts)
	if err != nil {
		return diag.Errorf("error creating Andromeda datacenter: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Datacenter == nil {
		return diag.Errorf("error creating Andromeda datacenter: empty response")
	}

	log.Printf("[DEBUG] Created Andromeda datacenter: %v", res)

	id := string(res.Payload.Datacenter.ID)
	d.SetId(id)

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := models.DatacenterProvisioningStatusACTIVE
	pending := models.DatacenterProvisioningStatusPENDINGCREATE
	datacenter, err = andromedaWaitForDatacenter(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetDatacenterResource(d, config, datacenter)

	return nil
}

func resourceSCIGSLBDatacenterV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Datacenters

	id := d.Id()
	datacenter, err := andromedaGetDatacenter(ctx, client, id)
	if err != nil {
		if _, ok := err.(*datacenters.GetDatacentersDatacenterIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	andromedaSetDatacenterResource(d, config, datacenter)

	return nil
}

func resourceSCIGSLBDatacenterV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Datacenters

	id := d.Id()
	datacenter := &models.Datacenter{}

	if d.HasChange("admin_state_up") {
		v := d.Get("admin_state_up").(bool)
		datacenter.AdminStateUp = &v
	}
	if d.HasChange("city") {
		v := d.Get("city").(string)
		datacenter.City = &v
	}
	if d.HasChange("continent") {
		v := d.Get("continent").(string)
		datacenter.Continent = &v
	}
	if d.HasChange("country") {
		v := d.Get("country").(string)
		datacenter.Country = &v
	}
	if d.HasChange("latitude") {
		v := d.Get("latitude").(float64)
		datacenter.Latitude = &v
	}
	if d.HasChange("longitude") {
		v := d.Get("longitude").(float64)
		datacenter.Longitude = &v
	}
	if d.HasChange("name") {
		v := d.Get("name").(string)
		datacenter.Name = &v
	}
	if d.HasChange("state_or_province") {
		v := d.Get("state_or_province").(string)
		datacenter.StateOrProvince = &v
	}
	if d.HasChange("service_provider") {
		v := d.Get("service_provider").(string)
		datacenter.Provider = v
	}
	if d.HasChange("scope") {
		v := d.Get("scope").(string)
		datacenter.Scope = &v
	}
	if d.HasChange("project_id") {
		v := d.Get("project_id").(string)
		datacenter.ProjectID = &v
	}

	opts := &datacenters.PutDatacentersDatacenterIDParams{
		Datacenter: datacenters.PutDatacentersDatacenterIDBody{
			Datacenter: datacenter,
		},
		DatacenterID: strfmt.UUID(id),
		Context:      ctx,
	}
	_, err = client.PutDatacentersDatacenterID(opts)
	if err != nil {
		return diag.Errorf("error updating Andromeda datacenter: %s", err)
	}

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutUpdate)
	target := models.DatacenterProvisioningStatusACTIVE
	pending := models.DatacenterProvisioningStatusPENDINGUPDATE
	datacenter, err = andromedaWaitForDatacenter(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetDatacenterResource(d, config, datacenter)

	return nil
}

func resourceSCIGSLBDatacenterV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Datacenters

	id := d.Id()
	opts := &datacenters.DeleteDatacentersDatacenterIDParams{
		DatacenterID: strfmt.UUID(id),
		Context:      ctx,
	}
	_, err = client.DeleteDatacentersDatacenterID(opts)
	if err != nil {
		if _, ok := err.(*datacenters.DeleteDatacentersDatacenterIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Andromeda datacenter: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := "DELETED"
	pending := models.DatacenterProvisioningStatusPENDINGDELETE
	_, err = andromedaWaitForDatacenter(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func andromedaWaitForDatacenter(ctx context.Context, client datacenters.ClientService, id, target, pending string, timeout time.Duration) (*models.Datacenter, error) {
	log.Printf("[DEBUG] Waiting for %s datacenter to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    []string{pending},
		Refresh:    andromedaGetDatacenterStatus(ctx, client, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	datacenter, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*datacenters.GetDatacentersDatacenterIDNotFound); ok && target == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s datacenter to become %s: %s", id, target, err)
	}

	return datacenter.(*models.Datacenter), nil
}

func andromedaGetDatacenterStatus(ctx context.Context, client datacenters.ClientService, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		datacenter, err := andromedaGetDatacenter(ctx, client, id)
		if err != nil {
			return nil, "", err
		}

		return datacenter, datacenter.ProvisioningStatus, nil
	}
}

func andromedaGetDatacenter(ctx context.Context, client datacenters.ClientService, id string) (*models.Datacenter, error) {
	opts := &datacenters.GetDatacentersDatacenterIDParams{
		DatacenterID: strfmt.UUID(id),
		Context:      ctx,
	}
	res, err := client.GetDatacentersDatacenterID(opts)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil || res.Payload.Datacenter == nil {
		return nil, fmt.Errorf("error reading Andromeda datacenter: empty response")
	}

	return res.Payload.Datacenter, nil
}

func andromedaSetDatacenterResource(d *schema.ResourceData, config *Config, datacenter *models.Datacenter) {
	_ = d.Set("admin_state_up", ptrValue(datacenter.AdminStateUp))
	_ = d.Set("city", ptrValue(datacenter.City))
	_ = d.Set("continent", ptrValue(datacenter.Continent))
	_ = d.Set("country", ptrValue(datacenter.Country))
	_ = d.Set("latitude", ptrValue(datacenter.Latitude))
	_ = d.Set("longitude", ptrValue(datacenter.Longitude))
	_ = d.Set("name", ptrValue(datacenter.Name))
	_ = d.Set("project_id", ptrValue(datacenter.ProjectID))
	_ = d.Set("service_provider", datacenter.Provider)
	_ = d.Set("scope", ptrValue(datacenter.Scope))
	_ = d.Set("state_or_province", ptrValue(datacenter.StateOrProvince))

	// computed
	_ = d.Set("provisioning_status", datacenter.ProvisioningStatus)
	_ = d.Set("meta", datacenter.Meta)
	_ = d.Set("created_at", datacenter.CreatedAt.String())
	_ = d.Set("updated_at", datacenter.UpdatedAt.String())

	_ = d.Set("region", GetRegion(d, config))
}
