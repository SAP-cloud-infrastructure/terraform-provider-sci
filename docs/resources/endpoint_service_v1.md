---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_endpoint_service_v1"
sidebar_current: "docs-sci-resource-endpoint-service-v1"
description: |-
  Manage an Archer endpoint service.
---

# sci\_endpoint\_service\_v1

Use this resource to create, manage, and delete an endpoint service within the
SAP Cloud Infrastructure environment.

## Example Usage

```hcl
resource "sci_endpoint_service_v1" "service_1" {
  availability_zone = "region1a"
  name              = "service1"
  ip_addresses      = ["192.168.1.1"]
  port              = 8080
  network_id        = "982d8699-b1a0-4933-8b65-7f1cd8a0f78b"
  visibility        = "private"
  tags              = ["tag1", "tag2"]
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to create the endpoint service. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `enabled` - (Optional) Specifies if the service is enabled. Defaults to
  `true`.

* `availability_zone` - (Optional) The availability zone in which to create the
  service. Changing this forces a new resource to be created

* `name` - (Optional) The name of the endpoint service.

* `description` - (Optional) A description of the endpoint service.

* `ip_addresses` - (Required) A list of IP addresses associated with the
  service.

* `port` - (Optional) The port on which the service is exposed. (Deprecated in
  favor of `ports`).

* `ports` - (Required) A list of ports on which the service is exposed.

* `network_id` - (Required) The network ID associated with the service.
  Changing this forces a new resource to be created.

* `project_id` - (Optional) The project ID associated with the service.
  Changing this forces a new resource to be created.

* `service_provider` - (Optional) The provider of the service (`tenant` or
  `cp`). Changing this forces a new resource to be created. Defaults to
  `tenant`.

* `proxy_protocol` - (Optional) Specifies if the proxy protocol is used.
  Defaults to `true`.

* `require_approval` - (Optional) Specifies if the service requires approval.
  Defaults to `true`.

* `visibility` - (Optional) The visibility of the service (`private` or
  `public`). Defaults to `private`.

* `tags` - (Optional) A list of tags assigned to the service.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the endpoint service.
* `host` - The host name of the service.
* `status` - The current status of the service.
* `created_at` - The timestamp when the service was created.
* `updated_at` - The timestamp when the service was last updated.

## Import

An Archer endpoint service can be imported using the `id`, e.g.

```shell
$ terraform import sci_endpoint_service_v1.service_1 069d36b0-125c-4c34-994d-849693805980
```
