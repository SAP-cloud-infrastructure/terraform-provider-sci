---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_bgpvpn_interconnection_v2"
sidebar_current: "docs-sci-resource-bgpvpn-interconnection-v2"
description: |-
  Manage a BGP VPN interconnection between regions.
---

# sci\_bgpvpn\_interconnection\_v2

Use this resource to manage a BGP VPN interconnection between two regions
within the SAP Cloud Infrastructure environment.

## Example Usage

```hcl
provider "openstack" {
  auth_url = "http://identity.region1/v3"
  region   = "region1"
  alias    = "region1"
}

provider "sci" {
  auth_url = "http://identity.region1/v3"
  region   = "region1"
  alias    = "region1"
}

provider "openstack" {
  auth_url = "https://identity.region2/v3"
  region   = "region2"
  alias    = "region2"
}

provider "sci" {
  auth_url = "https://identity.region2/v3"
  region   = "region2"
  alias    = "region2"
}

resource "openstack_bgpvpn_v2" "bgpvpn_1" {
  provider = openstack.region1

  name = "bgpvpn_1"
}

resource "openstack_bgpvpn_router_associate_v2" "router_associate_1" {
  provider = openstack.region1

  bgpvpn_id = openstack_bgpvpn_v2.bgpvpn_1.id
  router_id = "f490f377-fcd5-4ec7-ba47-e3787d2b7cec"
}

resource "sci_bgpvpn_interconnection_v2" "interconnection_1" {
  provider = sci.region1

  name               = "remote"
  local_resource_id  = openstack_bgpvpn_v2.bgpvpn_1.id
  remote_resource_id = openstack_bgpvpn_v2.bgpvpn_2.id
  remote_region      = "region2"
}

resource "openstack_bgpvpn_v2" "bgpvpn_2" {
  provider = openstack.region2

  name = "bgpvpn_2"
}

resource "openstack_bgpvpn_router_associate_v2" "router_associate_2" {
  provider = openstack.region2

  bgpvpn_id = openstack_bgpvpn_v2.bgpvpn_2.id
  router_id = "07d4ec98-f0eb-4cad-b5b0-518c133610ac"
}

resource "sci_bgpvpn_interconnection_v2" "interconnection_2" {
  provider = sci.region2

  name                      = "remote"
  local_resource_id         = openstack_bgpvpn_v2.bgpvpn_2.id
  remote_resource_id        = openstack_bgpvpn_v2.bgpvpn_1.id
  remote_region             = "region1"
  remote_interconnection_id = sci_bgpvpn_interconnection_v2.interconnection_1.id
}
```

## Argument Reference

* `region` - (Optional) The region in which to create the interconnection. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `name` - (Required) The name of the BGP VPN interconnection.

* `type` - (Optional) The type of the BGP VPN interconnection. Defaults to
 `bgpvpn`.

* `project_id` - (Optional) The ID of the project in which to create the
  BGP VPN interconnection. This field is computed if not set.

* `local_resource_id` - (Required) The ID of the local BGP VPN resource to be
  connected.

* `remote_resource_id` - (Required) The ID of the remote BGP VPN resource to be
  connected.

* `remote_region` - (Required) The region of the remote BGP VPN resource.

* `remote_interconnection_id` - (Optional) The ID of the remote BGP VPN
  interconnection to be linked with.

* `state` - (Optional) The desired state of the BGP VPN interconnection. Can be
  `WAITING_REMOTE`, `VALIDATING`, `VALIDATED`, `ACTIVE` or `TEARDOWN`. Setting
  the state is available only to cloud administrators.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the BGP VPN interconnection.
* `local_parameters` - The parameters of the local BGP VPN interconnection.
* `remote_parameters` - The parameters of the remote BGP VPN interconnection.

## Import

A BGP VPN interconnection can be imported using the `id`, e.g.

```
$ terraform import sci_bgpvpn_interconnection_v2.interconnection_1 8dff01b4-d6c2-4509-b872-5be4e93ad8ef
```
