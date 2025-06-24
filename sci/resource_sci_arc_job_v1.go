package sci

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/v2/arc/v1/jobs"
)

func resourceSCIArcJobV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIArcJobV1Create,
		ReadContext:   resourceSCIArcJobV1Read,
		DeleteContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics { return nil },

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			"to": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(1, 86400),
			},

			"execute": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"chef"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"script": {
							Type:          schema.TypeString,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"execute.0.tarball"},
							ValidateFunc:  validation.NoZeroValues,
						},

						"tarball": {
							Type:          schema.TypeList,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"execute.0.script"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validateURL,
									},

									"path": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.NoZeroValues,
									},

									"arguments": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},

									"environment": {
										Type:     schema.TypeMap,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},

			"chef": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"execute"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:          schema.TypeList,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"chef.0.zero"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"omnitruck_url": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validateURL,
									},

									"chef_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										Default:      "latest",
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},

						"zero": {
							Type:          schema.TypeList,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"chef.0.enable"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"run_list": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},

									"recipe_url": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validateURL,
									},

									"attributes": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateFunc:     validateJSONObject,
										DiffSuppressFunc: diffSuppressJSONObject,
										StateFunc:        normalizeJSONString,
									},

									"debug": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
									},

									"nodes": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateFunc:     validateJSONArray,
										DiffSuppressFunc: diffSuppressJSONArray,
										StateFunc:        normalizeJSONString,
									},

									"node_name": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},

									"omnitruck_url": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validateURL,
									},

									"chef_version": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  "latest",
									},
								},
							},
						},
					},
				},
			},

			// Computed attributes
			"agent": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"action": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"payload": {
				Type:     schema.TypeString,
				Computed: true,
				// Don't print the huge log during the terraform plan/apply
				Sensitive: true,
			},

			"agent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"sender": {
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

			"project": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"log": {
				Type:     schema.TypeString,
				Computed: true,
				// Don't print the huge log during the terraform plan/apply
				Sensitive: true,
			},

			"user": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"domain_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"roles": {
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

func resourceSCIArcJobV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	var agent, action, payload string

	if v, ok := getOkExists(d, "execute"); ok {
		agent = "execute"
		action, payload, err = arcSCIArcJobV1BuildPayload(v.([]any))
	}
	if v, ok := getOkExists(d, "chef"); ok {
		agent = "chef"
		action, payload, err = arcSCIArcJobV1BuildPayload(v.([]any))
	}
	if err != nil {
		return diag.Errorf("Failed to detect an agent: %v", err)
	}

	if len(agent) == 0 {
		return diag.Errorf("Failed to detect an agent")
	}

	if len(action) == 0 {
		return diag.Errorf("Failed to detect a %s action", agent)
	}

	if len(payload) == 0 {
		return diag.Errorf("Failed to build %s agent %s action payload", agent, action)
	}

	createOpts := jobs.CreateOpts{
		To:      d.Get("to").(string),
		Timeout: d.Get("timeout").(int),
		Agent:   agent,
		Action:  action,
		Payload: payload,
	}

	log.Printf("[DEBUG] sci_arc_job_v1 create options: %#v", createOpts)

	job, err := jobs.Create(ctx, arcClient, createOpts).Extract()
	if err != nil {
		return diag.Errorf("Error creating sci_arc_job_v1: %s", err)
	}

	d.SetId(job.RequestID)

	timeout := d.Timeout(schema.TimeoutCreate)
	target := []string{
		"complete",
		"failed",
	}
	pending := []string{
		"queued",
		"executing",
	}
	err = waitForArcJobV1(ctx, arcClient, job.RequestID, target, pending, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSCIArcJobV1Read(ctx, d, meta)
}

func resourceSCIArcJobV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	job, err := jobs.Get(ctx, arcClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(CheckDeleted(d, err, "Unable to retrieve sci_arc_job_v1"))
	}

	log := arcJobV1GetLog(ctx, arcClient, job.RequestID)

	execute, err := arcSCIArcJobV1FlattenExecute(job)
	if err != nil {
		return diag.Errorf("Error extracting execute payload for %s sci_arc_job_v1: %v", job.RequestID, err)
	}
	chef, err := arcSCIArcJobV1FlattenChef(job)
	if err != nil {
		return diag.Errorf("Error extracting chef payload for %s sci_arc_job_v1: %v", job.RequestID, err)
	}

	_ = d.Set("version", job.Version)
	_ = d.Set("sender", job.Sender)
	_ = d.Set("to", job.To)
	_ = d.Set("agent_id", job.To)
	_ = d.Set("timeout", job.Timeout)
	_ = d.Set("agent", job.Agent)
	_ = d.Set("action", job.Action)
	_ = d.Set("payload", job.Payload)
	_ = d.Set("execute", execute)
	_ = d.Set("chef", chef)
	_ = d.Set("status", job.Status)
	_ = d.Set("created_at", job.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", job.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("project", job.Project)
	_ = d.Set("user", flattenArcJobUserV1(job.User))
	_ = d.Set("log", string(log))

	_ = d.Set("region", GetRegion(d, config))

	return nil
}
