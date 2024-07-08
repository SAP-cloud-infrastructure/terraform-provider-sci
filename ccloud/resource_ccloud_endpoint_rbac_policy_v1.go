package ccloud

import (
	"context"
	"log"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/archer/client/rbac"
	"github.com/sapcc/archer/models"
)

func resourceCCloudEndpointRBACV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCCloudEndpointRBACV1Create,
		ReadContext:   resourceCCloudEndpointRBACV1Read,
		UpdateContext: resourceCCloudEndpointRBACV1Update,
		DeleteContext: resourceCCloudEndpointRBACV1Delete,
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
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"project",
				}, false),
			},

			// computed
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

func resourceCCloudEndpointRBACV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Rbac

	// Create the rbac
	req := &models.Rbacpolicy{
		ProjectID: models.Project(d.Get("project_id").(string)),
		ServiceID: ptr(strfmt.UUID(d.Get("service_id").(string))),
		Target:    d.Get("target").(string),
	}
	if v, ok := d.GetOk("target_type"); ok {
		req.TargetType = ptr(v.(string))
	}

	opts := &rbac.PostRbacPoliciesParams{
		Body:    req,
		Context: ctx,
	}
	res, err := client.PostRbacPolicies(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error creating Archer RBAC policy: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error creating Archer RBAC policy: empty response")
	}

	log.Printf("[DEBUG] Created Archer RBAC policy: %v", res)

	id := string(res.Payload.ID)
	d.SetId(id)

	archerSetRBACPolicyResource(d, config, res.Payload)

	return nil
}

func resourceCCloudEndpointRBACV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Rbac

	opts := &rbac.GetRbacPoliciesRbacPolicyIDParams{
		RbacPolicyID: strfmt.UUID(d.Id()),
		Context:      ctx,
	}
	res, err := client.GetRbacPoliciesRbacPolicyID(opts, c.authFunc())
	if err != nil {
		if _, ok := err.(*rbac.GetRbacPoliciesRbacPolicyIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading Archer RBAC policy: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error reading Archer RBAC policy: empty response")
	}

	log.Printf("[DEBUG] Read Archer RBAC policy: %v", res)

	archerSetRBACPolicyResource(d, config, res.Payload)

	return nil
}

func resourceCCloudEndpointRBACV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Rbac

	id := d.Id()
	rbacPolicy := &models.Rbacpolicycommon{
		ProjectID: models.Project(d.Get("project_id").(string)),
		Target:    ptr(d.Get("target").(string)),
	}

	if d.HasChange("target_type") {
		v := d.Get("target_type").(string)
		rbacPolicy.TargetType = &v
	}

	opts := &rbac.PutRbacPoliciesRbacPolicyIDParams{
		Body:         rbacPolicy,
		RbacPolicyID: strfmt.UUID(id),
		Context:      ctx,
	}
	res, err := client.PutRbacPoliciesRbacPolicyID(opts, c.authFunc())
	if err != nil {
		return diag.Errorf("error updating Archer RBAC policy: %s", err)
	}
	if res == nil || res.Payload == nil {
		return diag.Errorf("error updating Archer RBAC policy: empty response")
	}

	archerSetRBACPolicyResource(d, config, res.Payload)

	return nil
}

func resourceCCloudEndpointRBACV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.archerV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Archer client: %s", err)
	}
	client := c.Rbac

	id := d.Id()
	opts := &rbac.DeleteRbacPoliciesRbacPolicyIDParams{
		RbacPolicyID: strfmt.UUID(id),
		Context:      ctx,
	}
	_, err = client.DeleteRbacPoliciesRbacPolicyID(opts, c.authFunc())
	if err != nil {
		if _, ok := err.(*rbac.DeleteRbacPoliciesRbacPolicyIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Archer endpoint: %s", err)
	}

	return nil
}

func archerSetRBACPolicyResource(d *schema.ResourceData, config *Config, rbacPolicy *models.Rbacpolicy) {
	d.Set("service_id", ptrValue(rbacPolicy.ServiceID))
	d.Set("project_id", rbacPolicy.ProjectID)
	d.Set("target", rbacPolicy.Target)
	d.Set("target_type", rbacPolicy.TargetType)

	// computed
	d.Set("created_at", rbacPolicy.CreatedAt.String())
	d.Set("updated_at", rbacPolicy.UpdatedAt.String())

	d.Set("region", GetRegion(d, config))
}
