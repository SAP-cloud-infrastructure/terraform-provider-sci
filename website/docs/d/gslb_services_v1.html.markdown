---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_services_v1"
sidebar_current: "docs-ccloud-datasource-services-v1"
description: |-
  Get information about available services in your CCloud project. This data source is available only for OpenStack cloud admins.
---

# ccloud\_services\_v1

Use this data source to get information about available services in your CCloud project. This data source is available only for OpenStack cloud admins.

## Example Usage

```hcl
data "ccloud_services_v1" "services_1" {
  // No arguments
}
```

## Attributes Reference

The `services` attribute is a list of maps, where each map represents a service and contains the following keys:

- `heartbeat`: The heartbeat of the service.
- `host`: The host of the service.
- `id`: The unique identifier of the service.
- `metadata`: The metadata map associated with the service.
- `provider`: The provider of the service.
- `rpc_address`: The RPC address of the service.
- `type`: The type of the service.
- `version`: The version of the service.
