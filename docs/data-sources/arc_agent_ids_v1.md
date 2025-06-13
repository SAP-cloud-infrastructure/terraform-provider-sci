---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_arc_agent_ids_v1"
sidebar_current: "docs-sci-datasource-arc-agent-ids-v1"
description: |-
  Get a list of Arc Agent IDs.
---

# sci\_arc\_agent\_ids\_v1

Use this data source to get a list of Arc Agent IDs.

## Example Usage

```hcl
data "sci_arc_agent_ids_v1" "agent_ids_1" {
  filter = "@os = 'linux' AND @platform = 'ubuntu'"
}
```

## Argument Reference

* `region` - (Optional) The region in which to obtain the Arc client. If
   omitted, the `region` argument of the provider is used.

* `filter` - (Optional) The filter, used to filter the desired Arc agents.

## Attributes Reference

`id` is set to hash of the returned agents ID list. In addition, the following
attributes are exported:

* `region` - See Argument Reference above.
* `filter` - See Argument Reference above.
* `ids` - The list of Arc Agent IDs.
