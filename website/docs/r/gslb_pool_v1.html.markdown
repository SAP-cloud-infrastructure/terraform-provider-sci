---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_gslb_pool_v1"
sidebar_current: "docs-sci-resource-gslb-pool-v1"
description: |-
  Manage GSLB Pools
---

# sci\_gslb\_pool\_v1

This resource allows you to manage GSLB Pools.

## Example Usage

```hcl
resource "sci_gslb_pool_v1" "pool_1" {
  admin_state_up = true
  domains        = ["4da21196-4f20-48e6-aa56-42a567f40598"]
  name           = "pool1"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Andromeda client. If
  omitted, the `region` argument of the provider is used. Changing this creates
  a new pool.

* `admin_state_up` - (Optional) Specifies whether the pool is administratively
  up or down. Defaults to `true`.

* `domains` - (Optional) A list of UUIDs referencing the domain names
  associated with the pool.

* `name` - (Optional) The name of the pool.

* `project_id` - (Optional) The ID of the project this pool belongs to. This
  field is computed if not set. Changes to this field will trigger a new
  resource.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `members` - The list of members (servers) associated with this pool.
* `monitors` - The list of health monitors associated with this pool.
* `provisioning_status` -  The provisioning status of the pool.
* `status` - The operational status of the pool.
* `created_at` - The timestamp when the pool was created.
* `updated_at` - The timestamp when the pool was last updated.

## Import

Pools can be imported using the `id`, e.g.

```hcl
$ terraform import sci_gslb_pool_v1.pool_1 a4182fdb-a763-451e-8fd8-05f79d57128b
```
