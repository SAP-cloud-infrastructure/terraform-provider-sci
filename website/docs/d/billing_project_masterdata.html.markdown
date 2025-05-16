---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_billing_project_masterdata"
sidebar_current: "docs-sci-datasource-billing-project-masterdata"
description: |-
  Get information on the Billing Project Masterdata
---

# sci\_billing\_project\_masterdata

Use this data source to get the Billing Project Masterdata.

~> **Note:** You _must_ have admin privileges in your OpenStack cloud to use
this resource for other tenant projects.

## Example Usage

```hcl
data "sci_billing_project_masterdata" "masterdata" {
  project_id = "30dd31bcac8748daaa75720dab7e019a"
}

output "cost_object" {
  value = data.sci_billing_project_masterdata.masterdata.cost_object
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Billing client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `project_id` - (Optional) A project ID. Available only for users with an
  admin access. Defaults to the current project scope.

## Attributes Reference

In addition to arguments above, extra attributes are exported. Please refer
to the `sci_billing_project_masterdata` resource arguments and attributes
[documentation](../resources/billing_project_masterdata.html) for more information.
