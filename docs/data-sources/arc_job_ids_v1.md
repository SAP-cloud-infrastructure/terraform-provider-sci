---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_arc_job_ids_v1"
sidebar_current: "docs-sci-datasource-arc-job-ids-v1"
description: |-
  Get a list of Arc Job IDs.
---

# sci\_arc\_job\_ids\_v1

Use this data source to get a list of Arc Job IDs.

## Example Usage

```hcl
data "sci_arc_agent_v1" "agent_1" {
  filter  = "@metadata_name = 'hostname'"
}

data "sci_arc_job_ids_v1" "job_ids_1" {
  agent_id = data.sci_arc_agent_v1.agent_1.id
  agent    = "chef"
  action   = "zero"
}
```

## Argument Reference

* `region` - (Optional) The region in which to obtain the Arc client. If
   omitted, the `region` argument of the provider is used.

* `agent_id` - (Optional) The ID of the Arc agent.

* `timeout` - (Optional) The Arc job timeout in seconds. If specified,
  must be between 1 and 86400 seconds.

* `agent` - (Optional) The agent type, which executed the Arc job. Can either
  be `chef` or `execute`.

* `action` - (Optional) The Arc job action type. Can either be `script`, `zero`,
  `tarball` or `enable`.

* `status` - (Optional) The Arc job status. Can either be `queued`,
  `executing`, `failed`, `complete`.

## Attributes Reference

`id` is set to hash of the returned jobs ID list. In addition, the following
attributes are exported:

* `region` - See Argument Reference above.
* `agent_id` - See Argument Reference above.
* `timeout` - See Argument Reference above.
* `agent` - See Argument Reference above.
* `action` - See Argument Reference above.
* `status` - See Argument Reference above.
* `ids` - The list of Arc Job IDs.
