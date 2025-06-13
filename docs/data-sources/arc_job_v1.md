---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_arc_job_v1"
sidebar_current: "docs-sci-datasource-arc-job-v1"
description: |-
  Get information on an Arc Job.
---

# sci\_arc\_job\_v1

Use this data source to get the ID and other attributes of an Arc Job.

## Example Usage

### Get Arc Job using a Job ID

```hcl
resource "sci_automation_run_v1" "run_1" {
  automation_id = "123"
  selector      = "@metadata_name = 'hostname'"
}

data "sci_arc_job_v1" "job_1" {
  job_id = sci_automation_run_v1.run_1.jobs[0]
}
```

### Get Arc Job using a filter

```hcl
data "sci_arc_agent_v1" "agent_1" {
  filter  = "@metadata_name = 'hostname'"
}

data "sci_arc_job_v1" "job_1" {
  agent_id = data.sci_arc_agent_v1.agent_1.id
  timeout  = 3600
  agent    = "chef"
  action   = "zero"
  status   = "complete"
}
```

## Argument Reference

* `region` - (Optional) The region in which to obtain the Arc client. If
   omitted, the `region` argument of the provider is used.

* `job_id` - (Optional) The ID of the known Arc job. Conflicts with `agent_id`,
  `timeout`, `agent`, `action` and `status`.

* `agent_id` - (Optional) The ID of the known Arc agent. Conflicts with
  `job_id`.

* `timeout` - (Optional) The Arc job timeout in seconds. If specified,
  must be between 1 and 86400 seconds. Conflicts with `job_id`.

* `agent` - (Optional) The agent type, which executed the Arc job. Can either
  be `chef` or `execute`. Conflicts with `job_id`.

* `action` - (Optional) The Arc job action type. Can either be `script`, `zero`,
  `tarball` or `enable`. Conflicts with `job_id`.

* `status` - (Optional) The Arc job status. Can either be `queued`,
  `executing`, `failed`, `complete`. Conflicts with `job_id`.

## Attributes Reference

`id` is set to the ID of the found job. In addition, the following attributes
are exported:

* `region` - See Argument Reference above.
* `job_id` - See Argument Reference above.
* `agent_id` - See Argument Reference above.
* `timeout` - See Argument Reference above.
* `execute` - See Argument Reference in the `sci_arc_job_v1` resource
  documentation.
* `chef` - See Argument Reference in the `sci_arc_job_v1` resource
  documentation.
* `agent` - See Argument Reference above.
* `action` - See Argument Reference above.
* `status` - See Argument Reference above.
* `to` - A read-only alias to the `agent_id`.
* `payload` - The Arc job payload.
* `version` - The Arc job version.
* `sender` - The Arc job sender.
* `created_at` - The date the Arc job was created.
* `updated_at` - The date the Arc job was last updated.
* `project` - The parent Openstack project ID.
* `log` - The Arc job log.
* `user` - The user, who submitted the Arc job. The structure is described
   below.

The `user` attribute has fields below:

* `id` - The OpenStack user ID.

* `name` - The OpenStack user name.

* `domain_id` - The OpenStack domain ID.

* `domain_name` - The OpenStack domain name.

* `roles` - The list of the OpenStack user roles.
