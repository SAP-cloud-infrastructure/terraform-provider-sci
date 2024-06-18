---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_gslb_domain_v1"
sidebar_current: "docs-ccloud-resource-gslb-domain-v1"
description: |-
  Manage GSLB domains in your CCloud project
---

# ccloud\_gslb\_domain\_v1

This resource allows you to manage GSLB domains in your CCloud project.

## Example Usage

```hcl
resource "ccloud_gslb_domain_v1" "domain_1" {
  admin_state_up    = true
  fqdn              = "example.com"
  mode              = "ROUND_ROBIN"
  name              = "example-domain"
  project_id        = "your-project-id"
  service_provider  = "akamai"
  record_type       = "A"
  aliases           = ["www.example.com", "api.example.com"]
}
```

## Argument Reference

The following arguments are supported:

- `admin_state_up` (Optional): Specifies whether the domain is administratively up or down. Defaults to `true`.
- `fqdn` (Required): The fully qualified domain name managed by this GSLB domain.
- `mode` (Optional): The load balancing mode for the domain. Supported values are `ROUND_ROBIN`, `WEIGHTED`, `GEOGRAPHIC`, and `AVAILABILITY`. Defaults to `ROUND_ROBIN`.
- `name` (Optional): The name of the GSLB domain.
- `project_id` (Optional): The ID of the project this domain belongs to. This field is computed if not set. Changes to this field will trigger a new resource.
- `service_provider` (Optional): The service provider for the GSLB domain. Supported values are `akamai` and `f5`. Defaults to `akamai`.
- `record_type` (Optional): The type of DNS record for the domain. Supported values are `A`, `AAAA`, `CNAME`, and `MX`. Defaults to `A`.
- `aliases` (Optional): A list of aliases (additional domain names) that are managed by this GSLB domain.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `cname_target`: The CNAME target for the GSLB domain. This is computed based on the GSLB configuration.
- `provisioning_status`: The provisioning status of the domain.
- `status`: The operational status of the domain.
- `created_at`: The timestamp when the domain was created.
- `updated_at`: The timestamp when the domain was last updated.

## Import

Domains can be imported using the `domain_id`, e.g.

```hcl
$ terraform import ccloud_gslb_domain_v1.domain_1 f0f599a9-3a0d-4b4c-88d2-40c4fb071bba
```
