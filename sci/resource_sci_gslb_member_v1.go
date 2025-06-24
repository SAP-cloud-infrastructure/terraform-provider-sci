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
	"github.com/sapcc/andromeda/client/members"
	"github.com/sapcc/andromeda/models"
)

func resourceSCIGSLBMemberV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIGSLBMemberV1Create,
		ReadContext:   resourceSCIGSLBMemberV1Read,
		UpdateContext: resourceSCIGSLBMemberV1Update,
		DeleteContext: resourceSCIGSLBMemberV1Delete,
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
			"address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"datacenter_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// computed
			"provisioning_status": {
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

func resourceSCIGSLBMemberV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Members

	// Create the member
	adminStateUp := d.Get("admin_state_up").(bool)
	address := d.Get("address").(string)
	port := int64(d.Get("port").(int))
	member := &models.Member{
		Address:      &address,
		AdminStateUp: &adminStateUp,
		Port:         &port,
	}
	if v, ok := d.GetOk("datacenter_id"); ok && v != "" {
		v := strfmt.UUID(v.(string))
		member.DatacenterID = &v
	}
	if v, ok := d.GetOk("name"); ok && v != "" {
		member.Name = ptr(v.(string))
	}
	if v, ok := d.GetOk("project_id"); ok && v != "" {
		member.ProjectID = ptr(v.(string))
	}
	if v, ok := d.GetOk("pool_id"); ok && v != "" {
		v := strfmt.UUID(v.(string))
		member.PoolID = &v
	}

	opts := &members.PostMembersParams{
		Member: members.PostMembersBody{
			Member: member,
		},
		Context: ctx,
	}
	res, err := client.PostMembers(opts)
	if err != nil {
		return diag.Errorf("error creating Andromeda member: %s", err)
	}
	if res == nil || res.Payload == nil || res.Payload.Member == nil {
		return diag.Errorf("error creating Andromeda member: empty response")
	}

	log.Printf("[DEBUG] Created Andromeda member: %v", res)

	id := string(res.Payload.Member.ID)
	d.SetId(id)

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutCreate)
	target := models.MemberProvisioningStatusACTIVE
	pending := models.MemberProvisioningStatusPENDINGCREATE
	member, err = andromedaWaitForMember(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetMemberResource(d, config, member)

	return nil
}

func resourceSCIGSLBMemberV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Members

	id := d.Id()
	member, err := andromedaGetMember(ctx, client, id)
	if err != nil {
		if _, ok := err.(*members.GetMembersMemberIDNotFound); ok {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	andromedaSetMemberResource(d, config, member)

	return nil
}

func resourceSCIGSLBMemberV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Members

	id := d.Id()
	member := &models.Member{
		Address:      ptr(d.Get("address").(string)),
		DatacenterID: ptr(strfmt.UUID(d.Get("datacenter_id").(string))),
		Port:         ptr(int64(d.Get("port").(int))),
	}

	if d.HasChange("admin_state_up") {
		v := d.Get("admin_state_up").(bool)
		member.AdminStateUp = &v
	}
	if d.HasChange("name") {
		v := d.Get("name").(string)
		member.Name = &v
	}
	if d.HasChange("project_id") {
		v := d.Get("project_id").(string)
		member.ProjectID = &v
	}
	if d.HasChange("pool_id") {
		v := strfmt.UUID(d.Get("pool_id").(string))
		member.PoolID = &v
	}

	opts := &members.PutMembersMemberIDParams{
		Member: members.PutMembersMemberIDBody{
			Member: member,
		},
		MemberID: strfmt.UUID(id),
		Context:  ctx,
	}
	_, err = client.PutMembersMemberID(opts)
	if err != nil {
		return diag.Errorf("error updating Andromeda member: %s", err)
	}

	// waiting for ACTIVE status
	timeout := d.Timeout(schema.TimeoutUpdate)
	target := models.MemberProvisioningStatusACTIVE
	pending := models.MemberProvisioningStatusPENDINGUPDATE
	member, err = andromedaWaitForMember(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	andromedaSetMemberResource(d, config, member)

	return nil
}

func resourceSCIGSLBMemberV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	c, err := config.andromedaV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating Andromeda client: %s", err)
	}
	client := c.Members

	id := d.Id()
	opts := &members.DeleteMembersMemberIDParams{
		MemberID: strfmt.UUID(id),
		Context:  ctx,
	}
	_, err = client.DeleteMembersMemberID(opts)
	if err != nil {
		if _, ok := err.(*members.DeleteMembersMemberIDNotFound); ok {
			return nil
		}
		return diag.Errorf("error deleting Andromeda member: %s", err)
	}

	// waiting for DELETED status
	timeout := d.Timeout(schema.TimeoutDelete)
	target := "DELETED"
	pending := models.MemberProvisioningStatusPENDINGDELETE
	_, err = andromedaWaitForMember(ctx, client, id, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func andromedaWaitForMember(ctx context.Context, client members.ClientService, id, target, pending string, timeout time.Duration) (*models.Member, error) {
	log.Printf("[DEBUG] Waiting for %s member to become %s.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    []string{pending},
		Refresh:    andromedaGetMemberStatus(ctx, client, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	member, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if _, ok := err.(*members.GetMembersMemberIDNotFound); ok && target == "DELETED" {
			return nil, nil
		}
		return nil, fmt.Errorf("error waiting for %s member to become %s: %s", id, target, err)
	}

	return member.(*models.Member), nil
}

func andromedaGetMemberStatus(ctx context.Context, client members.ClientService, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		member, err := andromedaGetMember(ctx, client, id)
		if err != nil {
			return nil, "", err
		}

		return member, member.ProvisioningStatus, nil
	}
}

func andromedaGetMember(ctx context.Context, client members.ClientService, id string) (*models.Member, error) {
	opts := &members.GetMembersMemberIDParams{
		MemberID: strfmt.UUID(id),
		Context:  ctx,
	}
	res, err := client.GetMembersMemberID(opts)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Payload == nil || res.Payload.Member == nil {
		return nil, fmt.Errorf("error reading Andromeda member: empty response")
	}

	return res.Payload.Member, nil
}

func andromedaSetMemberResource(d *schema.ResourceData, config *Config, member *models.Member) {
	_ = d.Set("admin_state_up", ptrValue(member.AdminStateUp))
	_ = d.Set("address", ptrValue(member.Address))
	_ = d.Set("datacenter_id", ptrValue(member.DatacenterID))
	_ = d.Set("name", ptrValue(member.Name))
	_ = d.Set("pool_id", ptrValue(member.PoolID))
	_ = d.Set("port", ptrValue(member.Port))
	_ = d.Set("project_id", ptrValue(member.ProjectID))

	// computed
	_ = d.Set("provisioning_status", member.ProvisioningStatus)
	_ = d.Set("status", member.Status)
	_ = d.Set("created_at", member.CreatedAt.String())
	_ = d.Set("updated_at", member.UpdatedAt.String())

	_ = d.Set("region", GetRegion(d, config))
}
