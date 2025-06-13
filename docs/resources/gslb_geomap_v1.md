---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_gslb_geomap_v1"
sidebar_current: "docs-sci-resource-gslb-geomap-v1"
description: |-
  Manage GSLB geographical maps
---

# sci\_gslb\_geomap\_v1

~> **Note:** This kind of resources is experimental in Andromeda.

This resource allows to manage GSLB geographical maps.

## Example Usage

```hcl
resource "sci_gslb_geomap_v1" "geomap_1" {
  default_datacenter = "5c978d3c-a6c8-4322-9788-81a24212e958"
  name               = "geomap1"
  service_provider   = "akamai"
  scope              = "private"
  assignments {
    country    = "DE"
    datacenter = "5bfafa80-dbb9-4f7b-82a8-b60729373f5e"
  }
  assignments {
    country    = "US"
    datacenter = "e242ff7e-9f8f-4571-b8c7-82014ab6918c"
  }
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Andromeda client. If
  omitted, the `region` argument of the provider is used. Changing this creates
  a new geographical map.

* `default_datacenter` - (Required) The UUID of the default data center to use
  when no country match is found in the assignments.

* `name` - (Optional) The name of the geographical map.

* `project_id` - (Optional) The ID of the project this geographical map belongs
  to. This field is computed if not set. Changes to this field will trigger a
  new resource.

* `service_provider` - (Optional) The service provider for the GSLB geographical
  map. Supported values are `akamai` and `f5`. Defaults to `akamai`.

* `scope` - (Optional) The scope of the geographical map. Supported values are
  `private` and `shared`. Defaults to `private`.

* `assignments` - (Optional) A list of country to data center mappings. Each
  assignment specifies a `country` and a `datacenter` UUID.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` -  The ID of the geographical map.
* `provisioning_status` -  The provisioning status of the geographical map.
* `created_at` -  The timestamp when the geographical map was created.
* `updated_at` -  The timestamp when the geographical map was last updated.

## Import

Geographical map can be imported using the `id`, e.g.

```hcl
$ terraform import sci_gslb_geomap_v1.geomap_1 24404021-e95a-4362-af9c-0e0cf8c6b856
```
