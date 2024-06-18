---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_gslb_datacenter_v1"
sidebar_current: "docs-ccloud-resource-gslb-datacenter-v1"
description: |-
  Manage GSLB datacenters in your CCloud project
---

# ccloud\_gslb\_datacenter\_v1

This resource allows you to manage GSLB datacenters in your CCloud project.

## Example Usage

```hcl
resource "ccloud_gslb_datacenter_v1" "datacenter_1" {
  admin_state_up = true
  city = "City Name"
  continent = "Continent Name"
  country = "Country Name"
  latitude = 12.34
  longitude = 56.78
  name = "Datacenter Name"
  project_id = "Project ID"
  service_provider = "akamai"
  scope = "private"
  state_or_province = "State or Province Name"
}
```

## Argument Reference

The following arguments are supported:

- `admin_state_up` - (Optional) Whether the datacenter is administratively up. Default is `true`.
- `city` - (Optional) The city where the datacenter is located.
- `continent` - (Optional) The continent where the datacenter is located.
- `country` - (Optional) The country where the datacenter is located.
- `latitude` - (Optional) The latitude of the datacenter.
- `longitude` - (Optional) The longitude of the datacenter.
- `name` - (Optional) The name of the datacenter.
- `project_id` - (Optional) The ID of the project that the datacenter belongs to. Changes to this field will trigger a new resource.
- `service_provider` - (Optional) The service provider of the datacenter. Can be either `akamai` or `f5`. Default is `akamai`.
- `scope` - (Optional) The scope of the datacenter. Can be either `private` or `shared`. Default is `private`.
- `state_or_province` - (Optional) The state or province where the datacenter is located.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `provisioning_status` - The provisioning status of the datacenter.
- `meta` - Metadata associated with the datacenter.
- `created_at` - The time when the datacenter was created.
- `updated_at` - The time when the datacenter was last updated.

## Import

Datacenters can be imported using the `datacenter_id`, e.g.

```hcl
$ terraform import ccloud_gslb_datacenter_v1.datacenter_1 041053d5-e1ce-4724-bf96-aeeda1df2465
```
