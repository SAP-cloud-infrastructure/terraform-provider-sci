---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_endpoint_rbac_policy_v1"
sidebar_current: "docs-ccloud-resource-endpoint-rbac-policy-v1"
description: |-
  Manage RBAC policies for endpoints within the Converged Cloud environment.
---

# ccloud\_endpoint\_rbac\_policy\_v1

Use this resource to create, manage, and delete RBAC policies for endpoints within the Converged Cloud environment. This resource allows you to define access control policies for services and projects, specifying who can access what within your cloud environment.

## Example Usage

```hcl
resource "ccloud_endpoint_service_v1" "service_1" {
  name         = "svc1"
  port         = 80
  ip_addresses = ["192.168.1.2"]
  network_id   = "a7ec6c35-4e17-4e97-aa2b-0d93e56bb6c7"
}

resource "ccloud_endpoint_rbac_policy_v1" "rbac_1" {
  service_id = cc_endpoint_service_v1.service_1.id
  target     = "ea8e0fa95bc145cba3d58170d76f7643"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to create the RBAC policy. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `service_id` - (Required) The ID of the service to which the policy applies.

* `project_id` - (Optional) The ID of the project within which the policy is
  created. If omitted, the project ID of the provider is used.

* `target` - (Required) The ID of the target project to which the policy
  applies.

* `target_type` - (Optional) Specifies the type of the target. Valid values are
  `project`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the RBAC policy.
* `created_at` - The timestamp when the RBAC policy was created.
* `updated_at` - The timestamp when the RBAC policy was last updated.

## Import

Archer RBAC policies can be imported using the `id`, e.g.

```shell
$ terraform import ccloud_endpoint_rbac_policy_v1.rbac_1 b6e99485-bfcc-415f-acef-6ea2b3984c03
```
