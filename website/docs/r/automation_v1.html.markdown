---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_automation_v1"
sidebar_current: "docs-sci-resource-automation-v1"
description: |-
  Create a Lyra Automation.
---

# sci\_automation\_v1

Use this resource to create a Lyra Automation.

~> **Note:** All arguments and attributes, including repository credentials will
be stored in the raw state as plaintext.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

### Chef automation

```hcl
resource "sci_automation_v1" "chef_automation_1" {
  name            = "automation"
  repository      = "https://example.com/org/repo.git"
  type            = "Chef"
  run_list        = ["recipe[repo::default]"]
  chef_attributes = <<EOF
{"foo": "bar"}
EOF
}
```

### Script automation

```hcl
resource "sci_automation_v1" "script_automation_1" {
  name        = "automation"
  repository  = "https://example.com/org/repo.git"
  type        = "Script"
  path        = "nginx_install.sh"
  arguments   = ["--version", "1.15.9"]
  environment = {
    foo = "bar"
  }
}
```

### Using repository credentials

```hcl
resource "sci_automation_v1" "chef_automation_1" {
  name                   = "automation"
  repository             = "https://example.com/org/repo.git"
  repository_credentials = "githubToken"
  type                   = "Chef"
  run_list               = ["recipe[repo::default]"]
  chef_attributes        = <<EOF
{"foo": "bar"}
EOF
}
```

## Argument Reference

* `region` - (Optional) The region in which to obtain the Automation client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `name` - (Required) The name of the Lyra automation.

* `repository` - (Required) A valid URL to the automation repository.

* `type` - (Required) The type of the Lyra automation. Can either be `Script`
  or `Chef`. Changing this forces a new resource to be created.

* `repository_revision` - (Optional) The repository revision. Defaults to
  `master`.

* `repository_credentials` - (Optional) The credentials needed to access the
  repository (e.g.: git token or ssh key).

* `timeout` - (Optional) The automation timeout in seconds.

* `run_list` - (Required for the Chef type) An ordered list of Chef roles and/or
  recipes that are run in the exact order.

* `chef_attributes` - (Optional for the Chef type) A map of Chef cookbook
  attributes. Must be a valid JSON object.

* `debug` - (Optional for the Chef type) An enabled debug mode will not delete
  the temporary working directory on the instance when the automation job
  exists. Defaults to `false`.

* `chef_version` - (Optional for the Chef type) The Chef version to run the
  cookbook. Defaults to `latest`.

* `path` - (Required for the Script type) The Script path.

* `arguments` - (Optional for the Script type) The Script arguments list.

* `environment` - (Optional for the Script type) The Script environment map.

## Attributes Reference

`id` is set to the ID of the found automation. In addition, the following
attributes are exported:

* `region` - See Argument Reference above.
* `name` - See Argument Reference above.
* `repository` - See Argument Reference above.
* `repository_revision` - See Argument Reference above.
* `repository_authentication_enabled` - Set to true when a
  `repository_credentials` is set.
* `timeout` - See Argument Reference above.
* `type` - See Argument Reference above.
* `run_list` - See Argument Reference above.
* `chef_attributes` - See Argument Reference above.
* `debug` - See Argument Reference above.
* `chef_version` - See Argument Reference above.
* `path` - See Argument Reference above.
* `arguments` - See Argument Reference above.
* `environment` - See Argument Reference above.
* `created_at` - The date the Lyra automation was created.
* `updated_at` - The date the Lyra automation was last updated.
* `project_id` - The parent Openstack project ID.

## Import

An Automation can be imported using the `id`, e.g.

```
$ terraform import sci_automation_v1.chef_automation_1 3dcd5f53-ea76-43f8-bf80-36ddd6ddf5a2
```
