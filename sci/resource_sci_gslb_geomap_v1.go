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
	geomaps "github.com/sapcc/andromeda/client/geographic_maps"
	"github.com/sapcc/andromeda/models"
)

func resourceSCIGSLBGeoMapV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIGSLBGeoMapV1Create,
		ReadContext:   resourceSCIGSLBGeoMapV1Read,
		UpdateContext: resourceSCIGSLBGeoMapV1Update,
		DeleteContext: resourceSCIGSLBGeoMapV1Delete,
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
			"assignments": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"country": {
							Type:     schema.TypeString,
							Required: true,
						},
						"datacenter": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Optional: true,
			},
			"default_datacenter": {
				Type:     schema.TypeString,
				Required: true,
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

func resourceSCIGSLBGeoMapV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.GeographicMaps

	// Create the geomap
	defaultDatacenter := strfmt.UUID(d.Get("default_datacenter").(string))
	provider := d.Get("service_provider").(string)
	scope := d.Get("scope").(string)
	geomap := &models.Geomap{
		DefaultDatacenter: &defaultDatacenter,
		Provider:          provider,
		Scope:             &scope,
	}
	if v, ok := d.GetOk("name"); ok && v != "" {
		geomap.Name = ptr(v.(string))
	}
	if v, ok := d.GetOk("project_id"); ok && v != "" {
		geomap.ProjectID = ptr(v.(string))
	}
	if v, ok := d.GetOk("assignments"); ok {
		geomap.Assignments = andromedaExpandGeoMapAssignments(v.([]any))
	}

	opts := &geomaps.PostGeomapsParams{
		Geomap: geomaps.PostGeomapsBody{
			Geomap: geomap,
		},
		Context: ctx,
	}
	res, err := client.PostGeomaps(opts)
	if err != nil {
		return diag.Errorf("error creating Andromeda geographic map: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Geomap == nil {
		return diag.Errorf("error creating Andromeda geographic map: empty response")
	}

	log.Printf("[DEBUG] Created Andromeda geographic map: %vs", res)

	id := string(res.Payload.Geomap.ID)
	d.SetId(id)

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := models.GeomapProvisioningStatusACTIVE
	pending := models.GeomapProvisioningStatusPENDINGCREATE
	geomap, err = andromedaWaitForGeoMap(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetGeoMapResource(d, config, geomap)

	return nil
}

func resourceSCIGSLBGeoMapV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.GeographicMaps

	id := d.Id()
	geomap, err := andromedaGetGeoMap(ctx, client, id)
	if err != nil {
		if _, ok := err.(*geomaps.GetGeomapsGeomapIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	andromedaSetGeoMapResource(d, config, geomap)

	return nil
}

func resourceSCIGSLBGeoMapV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.GeographicMaps

	id := d.Id()
	geomap := &models.Geomap{
		DefaultDatacenter: ptr(strfmt.UUID(d.Get("default_datacenter").(string))),
	}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		geomap.Name = &v
	}
	if d.HasChange("project_id") {
		v := d.Get("project_id").(string)
		geomap.ProjectID = &v
	}
	if d.HasChange("service_provider") {
		geomap.Provider = d.Get("service_provider").(string)
	}
	if d.HasChange("scope") {
		v := d.Get("scope").(string)
		geomap.Scope = &v
	}
	if d.HasChange("assignments") {
		v := d.Get("assignments").([]any)
		geomap.Assignments = andromedaExpandGeoMapAssignments(v)
	}

	opts := &geomaps.PutGeomapsGeomapIDParams{
		GeomapID: strfmt.UUID(id),
		Geomap: geomaps.PutGeomapsGeomapIDBody{
			Geomap: geomap,
		},
		Context: ctx,
	}
	_, err = client.PutGeomapsGeomapID(opts)
	if err != nil {
		return diag.Errorf("error updating Andromeda geographic map: %s", err)
	}

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutUpdate)
	target := models.GeomapProvisioningStatusACTIVE
	pending := models.GeomapProvisioningStatusPENDINGUPDATE
	geomap, err = andromedaWaitForGeoMap(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetGeoMapResource(d, config, geomap)

	return nil
}

func resourceSCIGSLBGeoMapV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.GeographicMaps

	id := d.Id()
	opts := &geomaps.DeleteGeomapsGeomapIDParams{
		GeomapID: strfmt.UUID(id),
		Context:  ctx,
	}
	_, err = client.DeleteGeomapsGeomapID(opts)
	if err != nil {
		if _, ok := err.(*geomaps.DeleteGeomapsGeomapIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Andromeda geographic map: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := "DELETED"
	pending := models.GeomapProvisioningStatusPENDINGDELETE
	_, err = andromedaWaitForGeoMap(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func andromedaWaitForGeoMap(ctx context.Context, client geomaps.ClientService, id, target, pending string, timeout time.Duration) (*models.Geomap, error) {
	log.Printf("[DEBUG] Waiting for %s geographic map to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    []string{pending},
		Refresh:    andromedaGetGeoMapStatus(ctx, client, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	geomap, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*geomaps.GetGeomapsGeomapIDNotFound); ok && target == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s geographic map to become %s: %s", id, target, err)
	}

	return geomap.(*models.Geomap), nil
}

func andromedaGetGeoMapStatus(ctx context.Context, client geomaps.ClientService, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		geomap, err := andromedaGetGeoMap(ctx, client, id)
		if err != nil {
			return nil, "", err
		}

		return geomap, geomap.ProvisioningStatus, nil
	}
}

func andromedaGetGeoMap(ctx context.Context, client geomaps.ClientService, id string) (*models.Geomap, error) {
	opts := &geomaps.GetGeomapsGeomapIDParams{
		GeomapID: strfmt.UUID(id),
		Context:  ctx,
	}
	res, err := client.GetGeomapsGeomapID(opts)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil || res.Payload.Geomap == nil {
		return nil, fmt.Errorf("error reading Andromeda geographic map: empty response")
	}

	return res.Payload.Geomap, nil
}

func andromedaSetGeoMapResource(d *schema.ResourceData, config *Config, geomap *models.Geomap) {
	_ = d.Set("default_datacenter", geomap.DefaultDatacenter.String())
	_ = d.Set("name", ptrValue(geomap.Name))
	_ = d.Set("project_id", ptrValue(geomap.ProjectID))
	_ = d.Set("service_provider", geomap.Provider)
	_ = d.Set("scope", ptrValue(geomap.Scope))
	_ = d.Set("assignments", andromedaFlattenGeoMapAssignments(geomap.Assignments))

	// computed
	_ = d.Set("provisioning_status", geomap.ProvisioningStatus)
	_ = d.Set("created_at", geomap.CreatedAt.String())
	_ = d.Set("updated_at", geomap.UpdatedAt.String())

	_ = d.Set("region", GetRegion(d, config))
}

func andromedaFlattenGeoMapAssignments(assignments []*models.GeomapAssignmentsItems0) []map[string]string {
	res := make([]map[string]string, len(assignments))
	for i, assignment := range assignments {
		v := make(map[string]string)
		v["country"] = assignment.Country
		v["datacenter"] = assignment.Datacenter.String()
		res[i] = v
	}
	return res
}

func andromedaExpandGeoMapAssignments(v []any) []*models.GeomapAssignmentsItems0 {
	res := make([]*models.GeomapAssignmentsItems0, len(v))
	for i, v := range v {
		v := v.(map[string]any)
		country := v["country"].(string)
		datacenter := v["datacenter"].(string)
		res[i] = &models.GeomapAssignmentsItems0{
			Country:    country,
			Datacenter: strfmt.UUID(datacenter),
		}
	}
	return res
}
