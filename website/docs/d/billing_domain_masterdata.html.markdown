---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_billing_domain_masterdata"
sidebar_current: "docs-sci-datasource-billing-domain-masterdata"
description: |-
  Get information on the Billing Domain Masterdata
---

# sci\_billing\_domain\_masterdata

Use this data source to get the Billing Domain Masterdata.

~> **Note:** You _must_ have admin privileges in your OpenStack cloud to use
this resource.

## Example Usage

```hcl
data "sci_billing_domain_masterdata" "masterdata" {
  domain_id = "01482666f9004d4ea6b3458205642c30"
}

output "cost_object" {
  value = data.sci_billing_domain_masterdata.masterdata.cost_object
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Billing client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `domain_id` - (Optional) A domain ID. Defaults to the current domain scope.

## Attributes Reference

In addition to arguments above, extra attributes are exported. Please refer
to the `sci_billing_domain_masterdata` resource arguments and attributes
[documentation](../resources/billing_domain_masterdata.html) for more information.
