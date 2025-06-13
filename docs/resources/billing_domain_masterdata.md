---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_billing_domain_masterdata"
sidebar_current: "docs-sci-resource-billing-domain-masterdata"
description: |-
  Manages Billing Domain Masterdata
---

# sci\_billing\_domain\_masterdata

Manages Billing Domain masterdata.

~> **Note:** The `terraform destroy` command destroys the
`sci_billing_domain_masterdata` state, but not the actual billing domain
masterdata.

~> **Note:** You _must_ have admin privileges in your OpenStack cloud to use
this resource.

## Example Usage

```hcl
resource "sci_billing_domain_masterdata" "masterdata" {
  responsible_primary_contact_id    = "D123456"
  responsible_primary_contact_email = "mail@example.com"

  cost_object {
    projects_can_inherit = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Billing client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `domain_id` - (Optional) A domain ID. Defaults to the current domain scope.

* `domain_name` - (Optional) A human-readable name for the domain.

* `description` - (Optional) A domain description.

* `additional_information` - (Optional) Freetext field for additional
  information for domain.

* `responsible_primary_contact_id` - (Optional) SAP-User-Id of primary contact
  for the domain.

* `responsible_primary_contact_email` - (Optional) Email-address of primary
  contact for the domain.

* `cost_object` - (Optional) The cost object. The `cost_object` object structure
  is documented below.

* `collector` - (Optional) The Collector of the domain and subprojects.

The `cost_object` block supports:

* `projects_can_inherit` - (Optional) Set to true, if the costobject should be
  inheritable for subprojects.

* `name` - Name or ID of the costobject.

* `type` - Type of the costobject. Can either be `IO` (internal order), `CC`
  (cost center), `WBS` (Work Breakdown Structure element) or `SO` (sales order).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The domain ID.
* `created_at` - The date the Lyra automation was created.
* `changed_at` - The date the Lyra automation was last updated.
* `changed_by` - The OpenStack user ID of the user performed the last update.
* `is_complete` - True, if the given masterdata is complete. Otherwise false.
* `missing_attributes` - A human readable text, showing, what information is missing.

## Import

Billing Domain Masterdata can be imported with a `domain_id` argument, e.g.

```
$ terraform import sci_billing_domain_masterdata.demo 30dd31bcac8748daaa75720dab7e019a
```
