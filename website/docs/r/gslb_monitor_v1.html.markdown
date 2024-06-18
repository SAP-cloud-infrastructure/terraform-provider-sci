---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_gslb_monitor_v1"
sidebar_current: "docs-ccloud-resource-gslb-monitor-v1"
description: |-
  Manage GSLB Monitors in your CCloud project
---

# ccloud\_gslb\_monitor\_v1

This resource allows you to manage GSLB Monitors in your CCloud project.

## Example Usage

```hcl
resource "ccloud_gslb_monitor_v1" "monitor_1" {
  admin_state_up = true
  interval       = 10
  name           = "example-monitor"
  pool_id        = "your-pool-id"
  project_id     = "your-project-id"
  receive        = "HTTP 200 OK"
  send           = "/health"
  timeout        = 5
  type           = "HTTP"
}
```

## Argument Reference

- `admin_state_up` (Optional): Specifies whether the monitor is administratively up or down. Defaults to `true`.
- `interval` (Optional): The time, in seconds, between sending probes to members.
- `name` (Optional): The name of the monitor.
- `pool_id` (Optional): The ID of the pool that this monitor is associated with.
- `project_id` (Optional): The ID of the project this monitor belongs to. This field is computed if not set. Changes to this field will trigger a new resource.
- `receive` (Optional): The expected response text from the monitored resource. Valid only for `TCP` monitor types.
- `send` (Optional): The HTTP request method and path that the monitor sends to the monitored resource.
- `timeout` (Optional): Maximum time, in seconds, the monitor waits to receive a response from the monitored resource. This field is computed if not set.
- `type` (Optional): The type of monitor, which determines the method used to check the health of the monitored resource. Supported types are `ICMP`, `HTTP`, `HTTPS`, `TCP`, and `UDP`. Defaults to `ICMP`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `provisioning_status` - The provisioning status of the monitor.
- `created_at` - The time when the monitor was created.
- `updated_at` - The time when the monitor was last updated.

## Import

Monitors can be imported using the `monitor_id`, e.g.

```hcl
$ terraform import ccloud_gslb_monitor_v1.monitor_1 de731802-f092-496d-9508-9e02eb6ba0b1
```
