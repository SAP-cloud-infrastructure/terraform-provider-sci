---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_autoscaling_v1"
sidebar_current: "docs-sci-resource-autoscaling-v1"
description: |-
  Manages a Castellum autoscaling configuration for a project resource
---

# sci\_autoscaling\_v1

Manages a [Castellum](https://github.com/sapcc/castellum) autoscaling
configuration for a resource type within a project. Castellum monitors usage
metrics and automatically scales resource sizes up or down when configured
thresholds are crossed.

## Example Usage

### NFS share autoscaling with high and critical thresholds

```hcl
resource "sci_autoscaling_v1" "nfs" {
  resource_type = "nfs-shares"

  high_threshold {
    usage_percent = 80.0
    delay_seconds = 1800
  }

  critical_threshold {
    usage_percent = 95.0
  }

  size_constraints {
    minimum = 10
    maximum = 2000
  }

  size_steps {
    percent = 20.0
  }
}
```

### Server group autoscaling with multiple usage metrics

```hcl
resource "sci_autoscaling_v1" "server_group" {
  resource_type = "project-quota"

  low_threshold {
    metric        = "cores"
    usage_percent = 20.0
    delay_seconds = 3600
  }

  low_threshold {
    metric        = "ram"
    usage_percent = 20.0
    delay_seconds = 3600
  }

  high_threshold {
    metric        = "cores"
    usage_percent = 80.0
    delay_seconds = 1800
  }

  high_threshold {
    metric        = "ram"
    usage_percent = 80.0
    delay_seconds = 1800
  }

  critical_threshold {
    metric        = "cores"
    usage_percent = 95.0
  }

  critical_threshold {
    metric        = "ram"
    usage_percent = 95.0
  }

  size_constraints {
    minimum                  = 50
    maximum                  = 500
    minimum_free             = 10
    minimum_free_is_critical = true
  }

  size_steps {
    percent = 25.0
  }
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Castellum client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `project_id` - (Optional) The ID of the project to configure autoscaling for.
  If omitted, the project ID is derived from the provider token scope. Changing
  this forces a new resource to be created.

* `resource_type` - (Required) The Castellum resource type to manage, e.g.
  `nfs-shares` or `project-quota`. Changing this forces a new resource to be
  created.

* `low_threshold` - (Optional) One or more threshold blocks defining the low
  usage level that triggers a scale-down. The `low_threshold` block structure
  is documented below.

* `high_threshold` - (Optional) One or more threshold blocks defining the high
  usage level that triggers a scale-up. The `high_threshold` block structure
  is documented below.

* `critical_threshold` - (Optional) One or more threshold blocks defining the
  critical usage level that triggers an immediate scale-up. The
  `critical_threshold` block structure is documented below.

* `size_constraints` - (Optional) Constraints on the minimum and maximum sizes
  that Castellum is permitted to scale to. The `size_constraints` block
  structure is documented below.

* `size_steps` - (Optional) Controls how much the resource size changes on each
  scaling operation. The `size_steps` block structure is documented below.

The `low_threshold`, `high_threshold`, and `critical_threshold` blocks support:

* `metric` - (Optional) The usage metric this threshold applies to. Defaults to
  `singular` for resources with a single usage metric (e.g. NFS shares). For
  resources with multiple usage metrics (e.g. CPU and RAM), specify one block
  per metric with the corresponding metric name.

* `usage_percent` - (Required) The usage percentage at which this threshold is
  triggered.

* `delay_seconds` - (Optional) The number of seconds Castellum must observe the
  usage above this threshold continuously before acting. Defaults to `0`
  (immediate).

The `size_constraints` block supports:

* `minimum` - (Optional) The minimum size Castellum is permitted to scale down
  to. Defaults to `0` (no lower bound).

* `maximum` - (Optional) The maximum size Castellum is permitted to scale up to.
  When omitted, there is no upper bound.

* `minimum_free` - (Optional) The minimum amount of free capacity to maintain.
  Castellum will scale up to preserve at least this amount of headroom. Defaults
  to `0` (no free capacity floor).

* `minimum_free_is_critical` - (Optional) When `true`, falling below
  `minimum_free` is treated as a critical threshold rather than a high
  threshold. Defaults to `false`.

The `size_steps` block supports:

* `percent` - (Optional) The percentage of the current size to grow or shrink
  by on each scaling operation.

* `single` - (Optional) When `true`, Castellum scales by a single unit at a
  time instead of using percentage steps. Defaults to `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource identifier, formatted as `<project_id>/<resource_type>`.
* `region` - See Argument Reference above.
* `project_id` - The resolved project ID.

## Timeouts

`sci_autoscaling_v1` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts)
configuration options:

* `create` - (Default `30 minutes`) How long to wait for the autoscaling
  configuration to be created.
* `update` - (Default `30 minutes`) How long to wait for the autoscaling
  configuration to be updated.
* `delete` - (Default `30 minutes`) How long to wait for the autoscaling
  configuration to be deleted.

## Import

Autoscaling configurations can be imported using the project ID and resource
type (`<project_id>/<resource_type>`), e.g.

```shell
$ terraform import sci_autoscaling_v1.nfs abc123def456/nfs-shares
```
