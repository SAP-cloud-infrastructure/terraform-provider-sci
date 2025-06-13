---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_arc_agent_v1"
sidebar_current: "docs-sci-datasource-arc-agent-v1"
description: |-
  Get information on an Arc Agent.
---

# sci\_arc\_agent\_v1

Use this data source to get the ID and other attributes of an available Arc
Agent.

The resource can wait for an Arc Agent to be available (the Arc Agent bootstrap
takes time due to compute instance boot time and cloud-init execution delay)
within the `timeouts`
[nested](https://www.terraform.io/docs/configuration/resources.html#operation-timeouts)
block argument. The default read timeout is 0, what means don't wait.

## Example Usage

### Get an Arc Agent by an agent ID

```hcl
data "sci_arc_agent_v1" "agent_1" {
  agent_id = "72f50dc1-03c2-4177-9ffa-d75929734c0d"
}
```

### Find an Arc Agent with a filter

The example below will be completed once it finds the exact one agent
satisfying the specified filter.

```hcl
data "sci_arc_agent_v1" "agent_1" {
  filter  = "@metadata_name = 'hostname'"

  timeouts {
    read = "10m"
  }
}
```

## Argument Reference

* `region` - (Optional) The region in which to obtain the Arc client. If
   omitted, the `region` argument of the provider is used.

* `agent_id` - (Optional) The ID of the known Arc agent. Conflicts with
  `filter`.

* `filter` - (Optional) The filter, used to filter the desired Arc agent.
   Conflicts with `agent_id`.

## Attributes Reference

`id` is set to the ID of the found agent. In addition, the following attributes
are exported:

* `region` - See Argument Reference above.
* `agent_id` - See Argument Reference above.
* `filter` - See Argument Reference above.
* `timeout` - See Argument Reference above.
* `display_name` - The Arc agent display name.
* `project` - The Arc agent parent OpenStack project ID.
* `organization` - The Arc agent parent OpenStack domain ID.
* `created_at` - The date the Arc agent was created.
* `updated_at` - The date the Arc agent was last updated.
* `updated_with` - The registration ID, used to submit the latest update.
* `updated_by` - The type of the application, submitted the latest update.
* `all_tags` - The map of tags, assigned on the Arc agent.
* `facts` - The map of facts, submitted by the Arc agent.
* `facts_agents` - The map of agent types enabled on the Arc agent.
