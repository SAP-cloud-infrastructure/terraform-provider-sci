---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_automation_v1"
sidebar_current: "docs-sci-datasource-automation-v1"
description: |-
  Get information on a Lyra Automation.
---

# sci\_automation\_v1

Use this data source to get the ID and other attributes of a Lyra Automation.

## Example Usage

```hcl
data "sci_automation_v1" "automation_1" {
  name = "chef-automation"
}
```

## Argument Reference

* `region` - (Optional) The region in which to obtain the Automation client. If
   omitted, the `region` argument of the provider is used.

* `name` - (Optional) The name filter.

* `repository` - (Optional) The repository filter.

* `repository_revision` - (Optional) The repository revision filter.

* `timeout` - (Optional) The automation timeout filter in seconds.

* `type` - (Optional) The automation type filter. Can either be `Script` or
  `Chef`.

* `debug` - (Optional) The Chef debug flag filter.

* `chef_version` - (Optional) The Chef version filter.

* `path` - (Optional) The Script path filter.

## Attributes Reference

`id` is set to the ID of the found automation. In addition, the following
attributes are exported:

* `region` - See Argument Reference above.
* `name` - See Argument Reference above.
* `repository` - See Argument Reference above.
* `repository_revision` - See Argument Reference above.
* `repository_authentication_enabled` - Set to true when a
  `repository_credentials` is set.
* `timeout` - See Argument Reference above.
* `type` - See Argument Reference above.
* `run_list` - The Chef run list.
* `chef_attributes` - The Chef attributes.
* `debug` - See Argument Reference above.
* `chef_version` - See Argument Reference above.
* `path` - See Argument Reference above.
* `arguments` - The Script arguments list.
* `environment` - The Script environment map.
* `created_at` - The date the Lyra automation was created.
* `updated_at` - The date the Lyra automation was last updated.
* `project_id` - The parent Openstack project ID.
