---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_gslb_quota_v1"
sidebar_current: "docs-ccloud-resource-gslb-quota-v1"
description: |-
  Manage GSLB Quotas in your CCloud project
---

# ccloud\_gslb\_quota\_v1

This resource allows you to manage GSLB Quotas in your CCloud project.

~> **Note:** This resource can be used only by OpenStack cloud administrators.

~> **Note:** The `terraform destroy` command will reset all the quotas back to
zero.

## Example Usage

```hcl
resource "ccloud_gslb_quota_v1" "quota_1" {
  datacenter = 10
  domain = 20
  member = 30
  monitor = 40
  pool = 50
  project_id = "ea3b508ba36142d9888dc087b014ef78"
}
```

## Argument Reference

The following arguments are supported:

- `datacenter` - (Optional) The number of datacenters for the quota.
- `domain` - (Optional) The number of domains for the quota.
- `member` - (Optional) The number of members for the quota.
- `monitor` - (Optional) The number of monitors for the quota.
- `pool` - (Optional) The number of pools for the quota.
- `project_id` - (Required) The ID of the project that the quota belongs to. Changes to this field will trigger a new resource.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes
are exported:

- `in_use_datacenter` - The number of datacenters currently in use.
- `in_use_domain` - The number of domains currently in use.
- `in_use_member` - The number of members currently in use.
- `in_use_monitor` - The number of monitors currently in use.
- `in_use_pool` - The number of pools currently in use.

## Import

Quotas can be imported using the `project_id`, e.g.

```hcl
$ terraform import ccloud_gslb_quota_v1.quota_1 ea3b508ba36142d9888dc087b014ef78
```
