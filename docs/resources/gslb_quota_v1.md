---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_gslb_quota_v1"
sidebar_current: "docs-sci-resource-gslb-quota-v1"
description: |-
  Manage GSLB Quotas
---

# sci\_gslb\_quota\_v1

This resource allows you to manage GSLB Quotas.

~> **Note:** This resource can be used only by OpenStack cloud administrators.

~> **Note:** The `terraform destroy` command will reset all the quotas back to
zero.

## Example Usage

```hcl
resource "sci_gslb_quota_v1" "quota_1" {
  datacenter    = 10
  domain_akamai = 20
  domain_f5     = 20
  member        = 30
  monitor       = 40
  pool          = 50
  project_id    = "ea3b508ba36142d9888dc087b014ef78"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Andromeda client. If
  omitted, the `region` argument of the provider is used. Changing this creates
  a new quota.

* `project_id` - (Required) The ID of the project that the quota belongs to.
  Changes to this field will trigger a new resource.

* `datacenter` - (Optional) The number of datacenters for the quota.

* `domain` - (Deprecated) Use `domain_akamai` and `domain_f5` instead. The number of
  domains for the quota. This field is deprecated and will be removed in a future
  version.

* `domain_akamai` - (Optional) The number of Akamai domains for the quota.

* `domain_f5` - (Optional) The number of F5 domains for the quota.

* `member` - (Optional) The number of members for the quota.

* `monitor` - (Optional) The number of monitors for the quota.

* `pool` - (Optional) The number of pools for the quota.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes
are exported:

* `id` - The ID of the project.
* `in_use_datacenter` - The number of datacenters currently in use.
* `in_use_domain` - The number of domains currently in use. This is a deprecated
  field and will be removed in a future version. Use `in_use_domain_akamai` and
  `in_use_domain_f5` instead.
* `in_use_domain_akamai` - The number of Akamai domains currently in use.
* `in_use_domain_f5` - The number of F5 domains currently in use.
* `in_use_member` - The number of members currently in use.
* `in_use_monitor` - The number of monitors currently in use.
* `in_use_pool` - The number of pools currently in use.

## Import

Quotas can be imported using the project `id`, e.g.

```hcl
$ terraform import sci_gslb_quota_v1.quota_1 ea3b508ba36142d9888dc087b014ef78
```
