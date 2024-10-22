---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_gslb_member_v1"
sidebar_current: "docs-ccloud-resource-gslb-member-v1"
description: |-
  Manage GSLB Members
---

# ccloud\_gslb\_member\_v1

This resource allows you to manage GSLB Members.

## Example Usage

```hcl
resource "ccloud_gslb_member_v1" "member_1" {
  address        = "192.168.0.1"
  admin_state_up = true
  datacenter_id  = "datacenter-uuid"
  name           = "example-member"
  pool_id        = "pool-uuid"
  port           = 80
  project_id     = "your-project-id"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Andromeda client. If
  omitted, the `region` argument of the provider is used. Changing this creates
  a new member.

* `address` - (Required) The IP address of the member.

* `admin_state_up` - (Optional) Specifies whether the member is
  administratively up or down. Defaults to `true`.

* `datacenter_id` - (Optional) The UUID of the data center where the member is
  located.

* `name` - (Optional) The name of the GSLB member.

* `pool_id` - (Optional) The UUID of the GSLB pool to which the member belongs.

* `port` - (Required) The port on which the member is accepting traffic.

* `project_id` (Optional): The ID of the project this member belongs to. This
  field is computed if not set. Changes to this field will trigger a new
  resource.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` -  The ID of the member.
* `provisioning_status` -  The provisioning status of the member.
* `status` -  The operational status of the member.
* `created_at` -  The timestamp when the member was created.
* `updated_at` -  The timestamp when the member was last updated.

## Import

Members can be imported using the `id`, e.g.

```hcl
$ terraform import ccloud_gslb_member_v1.member_1 63c4c7fa-a90f-4fa1-8f21-ed8dbba6bc4b
```
