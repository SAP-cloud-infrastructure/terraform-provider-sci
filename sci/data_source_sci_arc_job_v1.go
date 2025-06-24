package sci

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/gophercloud-sapcc/v2/arc/v1/jobs"
)

func dataSourceSCIArcJobV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIArcJobV1Read,

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
										Type:     schema.TypeString,
										Computed: true,
									},

									"debug": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"nodes": {
										Type:     schema.TypeString,
										Computed: true,
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

func dataSourceSCIArcJobV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(ctx, GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	var job jobs.Job
	jobID := d.Get("job_id").(string)

	if len(jobID) > 0 {
		err = jobs.Get(ctx, arcClient, jobID).ExtractInto(&job)
		if err != nil {
			return diag.Errorf("Unable to retrieve %s sci_arc_job_v1: %v", jobID, err)
		}
	} else {
		// filter arc jobs by parameters
		jobs, err := arcSCIArcJobV1Filter(ctx, d, arcClient, "sci_arc_job_v1")
		if err != nil {
			return diag.FromErr(err)
		}

		if len(jobs) == 0 {
			return diag.Errorf("No sci_arc_job_v1 found")
		}

		if len(jobs) > 1 {
			return diag.Errorf("More than one sci_arc_job_v1 found (%d)", len(jobs))
		}

		job = jobs[0]
	}

	log := arcJobV1GetLog(ctx, arcClient, job.RequestID)

	execute, err := arcSCIArcJobV1FlattenExecute(&job)
	if err != nil {
		return diag.Errorf("Error extracting execute payload for %s sci_arc_job_v1: %v", job.RequestID, err)
	}
	chef, err := arcSCIArcJobV1FlattenChef(&job)
	if err != nil {
		return diag.Errorf("Error extracting chef payload for %s sci_arc_job_v1: %v", job.RequestID, err)
	}

	d.SetId(job.RequestID)
	_ = d.Set("version", job.Version)
	_ = d.Set("sender", job.Sender)
	_ = d.Set("job_id", job.RequestID)
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
