---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_arc_job_v1"
sidebar_current: "docs-sci-resource-arc-job-v1"
description: |-
  Create an Arc Job.
---

# sci\_arc\_job\_v1

Use this resource to schedule an Arc Job on the desired Arc Agent. The resource
will wait for the final Job status: `failed` or `complete`.

The `terraform destroy` command destroys the `sci_arc_job_v1` state, but not
the remote Arc Job object.

## Example Usage

### Execute a script

```hcl
data "sci_arc_agent_v1" "agent_1" {
  filter  = "@metadata_name = 'hostname'"

  timeouts {
    read = "10m"
  }
}

resource "sci_arc_job_v1" "job_1" {
  to = sci_arc_agent_v1.agent_1.id

  execute {
    script = <<EOF
echo "Script start"
for i in {1..10}; do
  echo $i
  sleep 1s
done
echo "Script done"
EOF
  }
}

output "job_status" {
  value = sci_arc_job_v1.job_1.status
}
```

### Enable Chef agent

```hcl
data "sci_arc_agent_v1" "agent_1" {
  filter  = "@metadata_name = 'hostname'"

  timeouts {
    read = "10m"
  }
}

resource "sci_arc_job_v1" "job_1" {
  to = sci_arc_agent_v1.agent_1.id

  chef {
    enable {}
  }
}

output "job_status" {
  value = sci_arc_job_v1.job_1.status
}
```

### Execute Chef Zero

```hcl
data "sci_arc_agent_v1" "agent_1" {
  filter  = "@metadata_name = 'hostname'"

  timeouts {
    read = "10m"
  }
}

resource "sci_arc_job_v1" "job_1" {
  to = sci_arc_agent_v1.agent_1.id

  chef {
    zero {
      run_list   = ["recipe[repo::default]"]
      recipe_url = "https://example.com/path/to/chef-zero-recipe.tgz"
      debug      = true

      attributes = <<EOF
{"foo": "bar"}
EOF
    }
  }
}

output "job_status" {
  value = sci_arc_job_v1.job_1.status
}
```

## Argument Reference

* `region` - (Optional) The region in which to obtain the Arc client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `to` - (Required) The ID of the Arc agent. Changing this forces a new
  resource to be created.

* `timeout` - (Optional) The Arc job timeout in seconds. If specified,
  must be between 1 and 86400 seconds. Defaults to 3600. Changing this forces a
  new resource to be created.

* `execute` - (Required) Execute a regular script or a binary from the remote
  tar archive. The `execute` object structure is documented below. Conflicts
  with `chef`. Changing this forces a new resource to be created.

* `chef` - (Required) Execute a Chef Zero automation. The `chef` object
  structure is documented below. Conflicts with `execute`. Changing this forces
  a new resource to be created.

* `triggers` - (Optional) A map of arbitrary strings that, when changed, will
  force the Arc Job to re-execute.

The `execute` block supports:

* `script` - (Required) The `script` payload. Conflicts with `tarball`. Changing
  this forces a new resource to be created.

* `tarball` - (Required) The `tarball` payload. The `tarball` object structure
  is documented below. Conflicts with `script`. Changing this forces a new
  resource to be created.

The `tarball` block supports:

* `url` - (Required) A valid URL to the tar archive. Changing this forces a new
  resource to be created.

* `path` - (Required) A path to the binary to execute within the tar archive
  directory structure. Changing this forces a new resource to be created.

* `arguments` - (Optional) A list of arguments to be passed to the binary,
  specified in the `path` argument. Changing this forces a new resource to be
  created.

* `environment` - (Optional) A map of environment variables to be set before
  executing the binary, specified in the `path` argument. Changing this forces a
  new resource to be created.

The `chef` block supports:

* `enable` - (Required) Generates the payload, which enables the Chef Agent on
  the Arc Agent. Conflicts with `enable`. Changing this forces a new resource to
  be created.

* `zero` - (Required) The Chef Zero payload. The `zero` object structure is
  documented below. Conflicts with `enable`. Changing this forces a new resource
  to be created.

The `enable` block supports:

* `omnitruck_url` - (Optional) The Chef repository URL containing the Chef
  binaries to download. Defaults to `https://www.chef.io/chef/metadata`. Read
  more on the Chef Docs [website](https://docs.chef.io/api_omnitruck.html).
  Changing this forces a new resource to be created.

* `chef_version` - (Optional) The Chef version to run the cookbook. Defaults to
  `latest`. Changing this forces a new resource to be created.

The `zero` block supports (you can find more documentation on the Chef
Docs [website](https://docs.chef.io)):

* `run_list` - (Required) An ordered list of Chef roles and/or recipes that are
  run in the exact order. Changing this forces a new resource to be created.

* `recipe_url` - (Required) A valid URL to the remote Chef cookbook tar archive.
  Changing this forces a new resource to be created.

* `attributes` - (Optional) A map of Chef cookbook attributes. Must be a valid
  JSON object. Changing this forces a new resource to be created.

* `debug` - (Optional) An enabled debug mode will not delete the temporary
  working directory on the instance when the automation job exists. Defaults to
  `false`. Changing this forces a new resource to be created.

* `nodes` - (Optional) A list of Chef node objects. Each object of this list
  will be written as a `nodes/%index%.json` file within the Chef cookbook
  directory. Must be a valid JSON array. Changing this forces a new resource to
  be created.

* `node_name` - (Optional) The Chef cookbook `node_name`. Defaults to the Arc
  Agent ID. Changing this forces a new resource to be created.

* `omnitruck_url` - (Optional) The Chef repository URL containing the Chef
  binaries to download. Defaults to `https://www.chef.io/chef/metadata`. Read
  more on the Chef Docs [website](https://docs.chef.io/api_omnitruck.html).
  Changing this forces a new resource to be created.

* `chef_version` - (Optional) The Chef version to run the cookbook. Defaults to
  `latest`. Changing this forces a new resource to be created.

## Attributes Reference

`id` is set to the job ID. In addition, the following attributes are exported:

* `region` - See Argument Reference above.
* `to` - See Argument Reference above.
* `timeout` - See Argument Reference above.
* `execute` - See Argument Reference above.
* `chef` - See Argument Reference above.
* `agent` - The agent type, which executed the Arc job. Can either be `chef` or
  `execute`.
* `action` - The Arc job action type. Can either be `script`, `zero`, `tarball`
  or `enable`.
* `payload` - The Arc job JSON payload.
* `agent_id` - A read-only alias to the `to` argument.
* `version` - The Arc job version.
* `sender` - The Arc job sender.
* `status` - The Arc job status. Can either be `queued`, `executing`, `failed`,
  `complete`.
* `created_at` - The date the Arc job was created.
* `updated_at` - The date the Arc job was last updated.
* `project` - The parent Openstack project ID.
* `log` - The Arc job log.
* `user` - The user, who submitted the Arc job. The structure is described
   below.

The `user` attribute has fields below:

* `id` - The OpenStack user ID.

* `name` - The OpenStack user name.

* `domain_id` - The OpenStack domain ID.

* `domain_name` - The OpenStack domain name.

* `roles` - The list of the OpenStack user roles.
