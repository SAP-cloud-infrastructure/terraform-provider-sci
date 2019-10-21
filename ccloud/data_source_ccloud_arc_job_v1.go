package ccloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/sapcc/gophercloud-arc/arc/v1/jobs"
)

func dataSourceCCloudArcJobV1() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCCloudArcJobV1Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"job_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"agent_id", "timeout", "agent", "action", "status"},
			},

			"agent_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"job_id"},
			},

			"timeout": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.IntBetween(1, 86400),
				ConflictsWith: []string{"job_id"},
			},

			"agent": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"chef", "execute",
				}, false),
				ConflictsWith: []string{"job_id"},
			},

			"action": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"script", "zero", "tarball", "enable",
				}, false),
				ConflictsWith: []string{"job_id"},
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"queued", "executing", "failed", "complete",
				}, false),
				ConflictsWith: []string{"job_id"},
			},

			"payload": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"execute": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"script": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tarball": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"path": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"arguments": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},

									"environment": {
										Type:     schema.TypeMap,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"chef": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"omnitruck_url": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"chef_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"zero": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"run_list": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},

									"recipe_url": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"attributes": {
										Type:      schema.TypeString,
										Computed:  true,
										StateFunc: normalizeJsonString,
									},

									"debug": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"nodes": {
										Type:      schema.TypeString,
										Computed:  true,
										StateFunc: normalizeJsonString,
									},

									"node_name": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"omnitruck_url": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"chef_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"to": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An alias to agent_id",
			},

			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"sender": {
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
				Type:      schema.TypeString,
				Computed:  true,
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

func dataSourceCCloudArcJobV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	var job jobs.Job
	jobID := d.Get("job_id").(string)

	if len(jobID) > 0 {
		err = jobs.Get(arcClient, jobID).ExtractInto(&job)
		if err != nil {
			return fmt.Errorf("Unable to retrieve %s ccloud_arc_job_v1: %s", jobID, err)
		}
	} else {
		// filter arc jobs by parameters
		jobs, err := arcCCloudArcJobV1Filter(d, arcClient, "ccloud_arc_job_v1")
		if err != nil {
			return err
		}

		if len(jobs) == 0 {
			return fmt.Errorf("No ccloud_arc_job_v1 found")
		}

		if len(jobs) > 1 {
			return fmt.Errorf("More than one ccloud_arc_job_v1 found (%d)", len(jobs))
		}

		job = jobs[0]
	}

	log := arcJobV1GetLog(arcClient, job.RequestID)

	execute, err := arcCCloudArcJobV1FlattenExecute(&job)
	if err != nil {
		return fmt.Errorf("Error extracting execute payload for %s ccloud_arc_job_v1: %s", job.RequestID, err)
	}
	chef, err := arcCCloudArcJobV1FlattenChef(&job)
	if err != nil {
		return fmt.Errorf("Error extracting chef payload for %s ccloud_arc_job_v1: %s", job.RequestID, err)
	}

	d.SetId(job.RequestID)
	d.Set("version", job.Version)
	d.Set("sender", job.Sender)
	d.Set("job_id", job.RequestID)
	d.Set("to", job.To)
	d.Set("agent_id", job.To)
	d.Set("timeout", job.Timeout)
	d.Set("agent", job.Agent)
	d.Set("action", job.Action)
	d.Set("payload", job.Payload)
	d.Set("execute", execute)
	d.Set("chef", chef)
	d.Set("status", job.Status)
	d.Set("created_at", job.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", job.UpdatedAt.Format(time.RFC3339))
	d.Set("project", job.Project)
	d.Set("user", flattenArcJobUserV1(job.User))
	d.Set("log", string(log))

	d.Set("region", GetRegion(d, config))

	return nil
}
