---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_endpoint_service_v1"
sidebar_current: "docs-sci-data-source-endpoint-service-v1"
description: |-
  Retrieve information about an Archer endpoint service.
---

# sci\_endpoint\_service\_v1

Use this data source to get information about an Archer endpoint service within
the SAP Cloud Infrastructure environment. This can be used to fetch details of a
specific service by various selectors like name, project ID, or tags.

## Example Usage

```hcl
data "sci_endpoint_service_v1" "service_1" {
  name        = "my-service"
  project_id  = "fa84c217f361441986a220edf9b1e337"
}

output "service_ip_addresses" {
  value = data.sci_endpoint_service_v1.service_1.ip_addresses
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to query for the endpoint service.
  If omitted, the `region` argument of the provider is used.

* `enabled` - (Optional) Filter services by their enabled status.

* `availability_zone` - (Optional) Filter services by their availability zone.

* `name` - (Optional) The name of the endpoint service.

* `description` - (Optional) A description of the endpoint service.

* `ip_addresses` - (Optional) A list of IP addresses associated with the service.

* `port` - (Optional) The port on which the service is exposed.

* `network_id` - (Optional) The network ID associated with the service.

* `project_id` - (Optional) The project ID associated with the service.

* `service_provider` - (Optional) Filter services by their provider (`tenant` or `cp`).

* `proxy_protocol` - (Optional) Filter services by their use of the proxy protocol.

* `require_approval` - (Optional) Filter services by whether they require approval.

* `visibility` - (Optional) Filter services by their visibility (`private` or `public`).

* `tags` - (Optional) A list of tags assigned to the service.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the found endpoint service.
* `all_ip_addresses` - A list of all IP addresses associated with the service.
* `all_tags` - A list of all tags assigned to the service.
* `host` - The host of the service owner.
* `status` - The current status of the service.
* `created_at` - The timestamp when the service was created.
* `updated_at` - The timestamp when the service was last updated.
