---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_gslb_services_v1"
sidebar_current: "docs-sci-datasource-gslb-services-v1"
description: |-
  Get information about available GSLB services.
---

# sci\_gslb\_services\_v1

Use this data source to get information about available GSLB services. This
data source is available only for OpenStack cloud admins.

## Example Usage

```hcl
data "sci_gslb_services_v1" "services_1" {
  // No arguments
}
```

## Attributes Reference

The following attributes are exported:

* `id` - The unique ID for the GSLB services generated from data hash.
* `region` -  The region in which the GSLB services are available.
* `services` -  A list of maps representing the available GSLB services.

The `services` attribute is a list of maps, where each map represents a service
and contains the following keys:

* `heartbeat` -  The heartbeat of the service.
* `host` -  The host of the service.
* `id` -  The unique identifier of the service.
* `metadata` -  The metadata map associated with the service.
* `provider` -  The provider of the service.
* `rpc_address` -  The RPC address of the service.
* `type` -  The type of the service.
* `version` -  The version of the service.
