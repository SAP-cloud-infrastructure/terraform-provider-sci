---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_endpoint_quota_v1"
sidebar_current: "docs-sci-resource-endpoint-quota-v1"
description: |-
  Manage quotas for Archer endpoints and services.
---

# sci\_endpoint\_quota\_v1

Use this resource to create, manage, and delete quotas for endpoints and services within the SAP Cloud Infrastructure environment.

~> **Note:** This resource can be used only by OpenStack cloud administrators.

~> **Note:** The `terraform destroy` command will reset all the quotas back to
zero.

## Example Usage

```hcl
resource "sci_endpoint_quota_v1" "quota_1" {
  project_id = "08c49418f7274a57864cd468ebbfb062"
  endpoint   = 10
  service    = 5
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to manage the quota. If omitted,
  the `region` argument of the provider is used. Changing this forces a new
  resource to be created.

* `endpoint` - (Optional) The quota for the number of endpoints. This is the
  maximum number of Archer endpoints that can be created.

* `service` - (Optional) The quota for the number of services. This is the
  maximum number of Archer services that can be created.

* `project_id` - (Required) The ID of the project for which to manage quotas.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the project.
* `in_use_endpoint` - The number of endpoints currently in use.
* `in_use_service` - The number of services currently in use.

## Import

A quota can be imported using the project `id`, e.g.

```shell
$ terraform import sci_endpoint_quota_v1.quota_1 08c49418f7274a57864cd468ebbfb062
```
