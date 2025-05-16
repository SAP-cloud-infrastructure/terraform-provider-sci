---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_endpoint_v1"
sidebar_current: "docs-sci-resource-endpoint-v1"
description: |-
  Manage an Archer endpoint.
---

# sci\_endpoint\_v1

Use this resource to create, manage, and delete an Archer endpoint within the
SAP Cloud Infrastructure environment.

## Example Usage

```hcl
data "sci_endpoint_service_v1" "service_1" {
  name    = "service_1"
  status  = "AVAILABLE"
  enabled = true

  availability_zone = "zone1"
}

resource "sci_endpoint_v1" "endpoint_1" {
  name        = "endpoint_1"
  service_id  = data.sci_endpoint_service_v1.service_1.id
  tags        = ["tag1", "tag2"]

  target {
    network = "49b6480b-24d3-4376-a4c9-aecbb89e16d9"
  }
}
```

## Argument Reference

* `region` - (Optional) The region in which to create the endpoint. If omitted,
  the `region` argument of the provider is used. Changing this forces a new
  resource to be created.

* `name` - (Optional) The name of the endpoint.

* `description` - (Optional) A description of the endpoint.

* `project_id` - (Optional) The ID of the project in which to create the
  endpoint. Changing this forces a new resource to be created.

* `service_id` - (Required) The ID of the service to which the endpoint is
  connected. Changing this forces a new resource to be created.

* `tags` - (Optional) A list of tags assigned to the endpoint.

* `target` - (Required) A block that defines the target of the endpoint.
  Changing this forces a new resource to be created. The block must contain one
  of the following arguments:

  * `network` - (Optional) The ID of the network to which the endpoint is
    connected. Conflicts with `subnet` and `port`.

  * `subnet` - (Optional) The ID of the subnet to which the endpoint is
    connected. Conflicts with `network` and `port`.

  * `port` - (Optional) The ID of the port to which the endpoint is connected.
    Conflicts with `network` and `subnet`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the endpoint.
* `ip_address` - The IP address assigned to the endpoint.
* `status` - The current status of the endpoint.
* `created_at` - The timestamp when the endpoint was created.
* `updated_at` - The timestamp when the endpoint was last updated.

## Import

An Archer endpoint can be imported using the `id`, e.g.

```shell
$ terraform import sci_endpoint_v1.endpoint_1 5f955108-5d6a-422d-b460-b7de087953b3
```
