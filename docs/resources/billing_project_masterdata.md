---
layout: "sci"
page_title: "SAP Cloud Infrastructure: sci_billing_project_masterdata"
sidebar_current: "docs-sci-resource-billing-project-masterdata"
description: |-
  Manages Billing Project Masterdata
---

# sci\_billing\_project\_masterdata

Manages Billing Project masterdata.

~> **Note:** The `terraform destroy` command destroys the
`sci_billing_project_masterdata` state, but not the actual billing project
masterdata.

~> **Note:** You _must_ have admin privileges in your OpenStack cloud to use
this resource for other tenant projects.

## Example Usage

```hcl
resource "sci_billing_project_masterdata" "masterdata" {
  responsible_primary_contact_id    = "D123456"
  responsible_primary_contact_email = "mail@example.com"

  number_of_endusers = 100

  cost_object {
    inherited = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Billing client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `project_id` - (Optional) A project ID. Available only for users with an
  admin access. Defaults to the current project scope.

* `project_name` - (Optional) A human-readable name for the project. Available
  only for users with an admin access.

* `domain_id` - (Optional) A domain ID in which the project is contained.
  Available only for users with an admin access.

* `domain_name` - (Optional) A domain name in which the project is contained.
  Available only for users with an admin access.

* `parent_id` - (Optional) A project parent ID. Available only for users with
  an admin access.

* `project_type` - (Optional) A project type. Available only for users with an
  admin access.

* `description` - (Optional) A project description.

* `revenue_relevance` - (Optional) Indicating if the project is directly or
  indirectly creating revenue. Can either be `generating`", `enabling` or
  `other`.

* `business_criticality` - (Optional) Indicates how important the project for
  the business is. Can either be `dev`, `test` or `prod`.

* `number_of_endusers` - (Optional) An amount of end users. `-1` indicates that
  it is infinite.

* `additional_information` - (Optional) Freetext field for additional
  information for project.

* `responsible_primary_contact_id` - (Optional) SAP-User-Id of primary contact
  for the project.

* `responsible_primary_contact_email` - (Optional) Email-address of primary
  contact for the project.

* `responsible_operator_id` - (Optional) SAP-User-Id of the person who is
  responsible for operating the project.

* `responsible_operator_email` - (Optional) Email-address or DL of the
  person/group who is operating the project.

* `responsible_inventory_role_id` - (Optional) SAP-User-Id of the Person/entity
  responsible to correctly maintain assets in SAP's Global DC HW asset inventory
  SISM/CCIR.

* `responsible_inventory_role_email` - (Optional) Email-address or DL of the
  Person/entity responsible to correctly maintain assets in SAP's Global DC HW
  asset inventory SISM/CCIR.

* `responsible_infrastructure_coordinator_id` - (Optional) SAP-User-Id of the
  infrastructure coordinator.

* `responsible_infrastructure_coordinator_email` - (Optional) Email-address or
  DL of the infrastructure coordinator.

* `cost_object` - (Optional) The cost object. The `cost_object` object structure
  is documented below.

* `environment` - (Optional) Build environment of the project. Can either be
  `Prod`, `QA`, `Admin`, `DEV`, `Demo`, `Train`, `Sandbox`, `Lab` or `Test`.

* `soft_license_mode` - (Optional) Software License Mode. Can either be
  `Revenue Generating`, `Training & Demo`, `Development`, `Test & QS`,
  `Administration`, `Make`, `Virtualization-Host` or `Productive`.

* `type_of_data` - (Optional) Input parameter for KRITIS flag in CCIR. Can
  either be `SAP Business Process`, `Customer Cloud Service`, `Customer Business
  Process` or `Training & Demo Cloud`.

* `gpu_enabled` - (Optional) Indicates whether the project uses GPUs.

* `contains_pii_dpp_hr` - (Optional) Indicates whether the project contains
  sensitive personal data.

* `contains_external_customer_data` - (Optional) Indicates whether the project
  contains data from external customer.

* `ext_certification` - (Optional) Contains information about whether there is
  any external certification present in this project. The `ext_certification`
  object structure is documented below.

The `cost_object` block supports:

* `inherited` - (Optional) Shows, if the cost object is inherited. Required, if
  name/type not set.

* `name` - Name or ID of the costobject. Required, if `inherited` not true.

* `type` - Type of the costobject. Can either be `IO` (internal order), `CC`
  (cost center), `WBS` (Work Breakdown Structure element) or `SO` (sales order).
  Required, if `inherited` not true.

The `ext_certification` block supports boolean values indicating whether the
project has a corresponding certification:

* `c5` - C5 is a government-backed verification framework implemented by the
  German Federal Office for Information Security (BSI).

* `iso` - An ISO certification describes the process that confirms that ISO
  standards are being followed.

* `pci` - PCI certification ensures the security of card data at your business
  through a set of requirements established by the PCI SSC.

* `soc1` - SOC is a type of audit report that attests to the trustworthiness of
  services provided by a service organization.

* `soc2` - SOC is a type of audit report that attests to the trustworthiness of
  services provided by a service organization.

* `sox` - The law mandates strict reforms to improve financial disclosures from
  corporations and prevent accounting fraud.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The project ID.
* `created_at` - The date the Lyra automation was created.
* `changed_at` - The date the Lyra automation was last updated.
* `changed_by` - The OpenStack user ID of the user performed the last update.
* `is_complete` - True, if the given masterdata is complete. Otherwise false.
* `missing_attributes` - A human readable text, showing, what information is missing.
* `collector` - The Collector of the project.

## Import

Billing Project Masterdata can be imported with a `project_id` argument, e.g.

```
$ terraform import sci_billing_project_masterdata.demo 30dd31bcac8748daaa75720dab7e019a
```
