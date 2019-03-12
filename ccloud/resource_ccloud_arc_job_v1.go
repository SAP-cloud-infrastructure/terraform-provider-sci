package ccloud

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/kayrus/gophercloud-arc/arc/v1/jobs"
)

func resourceCCloudArcJobV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceCCloudArcJobV1Create,
		Read:   resourceCCloudArcJobV1Read,
		Delete: func(*schema.ResourceData, interface{}) error { return nil },

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
				ConflictsWith: []string{"chef_zero"},
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

			"chef_zero": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"execute"},
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
							ValidateFunc:     validateJsonObject,
							DiffSuppressFunc: diffSuppressJsonObject,
							StateFunc:        normalizeJsonString,
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
							ValidateFunc:     validateJsonArray,
							DiffSuppressFunc: diffSuppressJsonArray,
							StateFunc:        normalizeJsonString,
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
				MaxItems: 1,
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

func resourceCCloudArcJobV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	var agent string
	var action string
	var payload string

	if v, ok := d.GetOkExists("execute"); ok {
		agent = "execute"
		action, payload = arcCCloudArcJobV1BuildExecutePayload(v.([]interface{}))
		if len(action) == 0 {
			return fmt.Errorf("Failed to detect an execute action")
		}
	}

	if v, ok := d.GetOkExists("chef_zero"); ok {
		agent = "chef"
		action = "zero"
		payload = arcCCloudArcJobV1BuildChefPayload(v.([]interface{}))
	}

	if len(agent) == 0 {
		return fmt.Errorf("Failed to detect an agent")
	}

	if len(action) == 0 {
		return fmt.Errorf("Failed to detect an execute action")
	}

	if len(payload) == 0 {
		return fmt.Errorf("Failed to build %s agent %s action payload", agent, action)
	}

	createOpts := jobs.CreateOpts{
		To:      d.Get("to").(string),
		Timeout: d.Get("timeout").(int),
		Agent:   agent,
		Action:  action,
		Payload: payload,
	}

	log.Printf("[DEBUG] ccloud_arc_job_v1 create options: %#v", createOpts)

	job, err := jobs.Create(arcClient, createOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error creating ccloud_arc_job_v1: %s", err)
	}

	d.SetId(job.RequestID)

	timeout := d.Timeout(schema.TimeoutCreate)
	target := []string{"complete", "failed"}
	pending := []string{"queued", "executing"}
	err = waitForArcJobV1(arcClient, job.RequestID, target, pending, timeout)
	if err != nil {
		return err
	}

	return resourceCCloudArcJobV1Read(d, meta)
}

func resourceCCloudArcJobV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	job, err := jobs.Get(arcClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Unable to retrieve ccloud_arc_job_v1")
	}

	res := jobs.GetLog(arcClient, job.RequestID)
	if res.Err != nil {
		return fmt.Errorf("Error retrieving logs for %s ccloud_arc_job_v1: %s", job.RequestID, res.Err)
	}
	log, err := res.ExtractContent()
	if err != nil {
		return fmt.Errorf("Error extracting logs for %s ccloud_arc_job_v1: %s", job.RequestID, err)
	}

	execute, err := arcCCloudArcJobV1FlattenExecute(job)
	if err != nil {
		return fmt.Errorf("Error extracting execute payload for %s ccloud_arc_job_v1: %s", job.RequestID, err)
	}
	chefZero, err := arcCCloudArcJobV1FlattenChef(job)
	if err != nil {
		return fmt.Errorf("Error extracting chef payload for %s ccloud_arc_job_v1: %s", job.RequestID, err)
	}

	d.Set("version", job.Version)
	d.Set("sender", job.Sender)
	d.Set("to", job.To)
	d.Set("agent_id", job.To)
	d.Set("timeout", job.Timeout)
	d.Set("agent", job.Agent)
	d.Set("action", job.Action)
	d.Set("payload", job.Payload)
	d.Set("execute", execute)
	d.Set("chef_zero", chefZero)
	d.Set("status", job.Status)
	d.Set("created_at", job.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", job.UpdatedAt.Format(time.RFC3339))
	d.Set("project", job.Project)
	d.Set("user", flattenArcJobUserV1(job.User))
	d.Set("log", string(log))

	d.Set("region", GetRegion(d, config))

	return nil
}
