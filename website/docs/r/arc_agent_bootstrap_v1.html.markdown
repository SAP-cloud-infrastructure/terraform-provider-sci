---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_arc_agent_bootstrap_v1"
sidebar_current: "docs-sci-resource-arc-agent-bootstrap-v1"
description: |-
  Get the bootstrap information for a new Arc Agent.
---

# sci\_arc\_agent\_bootstrap\_v1

Use this resource to get the initialize data for a new Arc Agent. This data
could be used as an OpenStack user data, executed by cloud-init system on
compute instance boot.

The `terraform destroy` command destroys the `sci_arc_agent_bootstrap_v1`
state, but not the remote object, since the bootstrap data is a PKI token, not a
real resource.

The `terraform refresh` command doesn't refresh the bootstrap data, but reads it
from the Terraform state.

## Example Usage

### Get an Arc Agent bootstrap script for Linux cloud-init

```hcl
resource "sci_arc_agent_bootstrap_v1" "agent_1" {}

resource "openstack_compute_instance_v2" "node" {
  name        = "linux-vm"
  image_name  = "ubuntu-16.04-amd64"
  flavor_name = "m1.small"
  user_data   = sci_arc_agent_bootstrap_v1.agent_1.user_data

  network {
    name = "private_network"
  }
}
```

### Get an Arc Agent bootstrap script for Windows cloud-init

```hcl
resource "sci_arc_agent_bootstrap_v1" "agent_1" {
  type = "windows"
}

resource "openstack_compute_instance_v2" "node" {
  name        = "win-vm"
  image_name  = "windows-2016-amd64"
  flavor_name = "m1.large"
  user_data   = sci_arc_agent_bootstrap_v1.agent_1.user_data

  network {
    name = "private_network"
  }
}
```

### Get an extended Arc Agent bootstrap script for Windows cloud-init

```hcl
resource "sci_arc_agent_bootstrap_v1" "agent_1" {
  type = "json"
}

data "template_file" "user_data" {
  template = <<EOF
#ps1_sysnative
mkdir C:\monsoon\arc
(New-Object System.Net.WebClient).DownloadFile('$${update_url}/arc/windows/amd64/latest','C:\monsoon\arc\arc.exe')
C:\monsoon\arc\arc.exe init --endpoint $${endpoint_url} --update-uri $${update_url} --registration-url $${reg_url}

net user Administrator $${admin_passwd}
EOF

  vars = {
    token        = lookup(sci_arc_agent_bootstrap_v1.agent_1.raw_map, "token")
    reg_url      = lookup(sci_arc_agent_bootstrap_v1.agent_1.raw_map, "url")
    endpoint_url = lookup(sci_arc_agent_bootstrap_v1.agent_1.raw_map, "endpoint_url")
    update_url   = lookup(sci_arc_agent_bootstrap_v1.agent_1.raw_map, "update_url")
    admin_passwd = var.password
  }
}

resource "openstack_compute_instance_v2" "node" {
  name        = "win-vm"
  image_name  = "windows-2016-amd64"
  flavor_name = "m1.large"
  user_data   = data.template_file.user_data.rendered

  network {
    name = "private_network"
  }
}
```

## Argument Reference

* `region` - (Optional) The region in which to obtain the Arc client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `type` - (Optional) The bootstrap script type. Can either be `linux`,
  `windows`, `cloud-config` or `json`. Defaults to `cloud-config`. When `type`
  is set to `json`, an additional `raw_map` attribute is available with the
  decoded JSON response. Changing this forces a new resource to be created.

* `triggers` - (Optional) A map of arbitrary strings that, when changed, will
  force a new Arc PKI token to be issued.

## Attributes Reference

`id` is set to hash of the returned `user_data` content. In addition, the
following attributes are exported:

* `region` - See Argument Reference above.
* `type` - See Argument Reference above.
* `user_data` - The user data content, returned by the Arc server.
* `raw_map` - A map with the decoded JSON, when the `json` type is specified.
