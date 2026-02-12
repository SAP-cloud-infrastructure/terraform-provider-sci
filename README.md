SAP Cloud Infrastructure - Terraform Provider
=======================================================

Documentation: [registry.terraform.io](https://registry.terraform.io/providers/SAP-cloud-infrastructure/sci/latest/docs)

Maintainers
-----------

This provider plugin is maintained by:

  * [@kayrus](https://github.com/kayrus)
  * [@bugroger](https://github.com/BugRoger)

Requirements
------------

- [Terraform](https://developer.hashicorp.com/terraform/install) 1.x
- [Go](https://go.dev/doc/install) 1.26 (to build the provider plugin)

Usage
---------------------

The SAP Cloud Infrastructure provider is an extension to the [OpenStack Terraform
Provider](https://github.com/terraform-provider-openstack/terraform-provider-openstack).
It provides resources that allow to use Terraform for SAP Cloud Infrastructure's
additional services:

  * Kubernikus (Kubernetes as a Service)
  * Arc for Arc resources management
  * Lyra for Automation management
  * Billing for Billing management
  * Andromeda for GSLB / GTM (Global Server Load Balancing / Global Traffic Management)
  * Archer for Endpoint Services

The provider needs to be configured with the proper OpenStack credentials
before it can be used. For details see the OpenStack provider.

Building The Provider
---------------------

Clone the repository

```sh
$ git clone git@github.com:SAP-cloud-infrastructure/terraform-provider-sci
```

Enter the provider directory and build the provider

```sh
$ cd terraform-provider-sci
$ make build
```

Installing the provider
-----------------------

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```hcl
terraform {
  required_providers {
    sci = {
      source = "SAP-cloud-infrastructure/sci"
    }
  }
}

provider "sci" {
  # Configuration options
}
```

Using the provider
----------------------
Please see the documentation at [registry.terraform.io](https://registry.terraform.io/providers/SAP-cloud-infrastructure/sci/latest/docs).

Or you can browse the documentation within this repo [here](https://github.com/SAP-cloud-infrastructure/terraform-provider-sci/tree/master/website/docs).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](https://go.dev/) installed on your machine (version 1.26+ is *required*).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the current directory.

```sh
$ make build
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

Releasing the Provider
----------------------

This repository contains a GitHub Action configured to automatically build and
publish assets for release when a tag is pushed that matches the pattern `v*`
(ie. `v0.1.0`).

A [Gorelaser](https://goreleaser.com/) configuration is provided that produce
build artifacts matching the [layout required](https://developer.hashicorp.com/terraform#manually-preparing-a-release)
to publish the provider in the Terraform Registry.

Releases will as drafts. Once marked as published on the GitHub Releases page,
they will become available via the Terraform Registry.

Support, Feedback, Contributing
-------------------------------

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-an-issue). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](https://github.com/SAP-cloud-infrastructure/.github/blob/main/CONTRIBUTING.md).

Security / Disclosure
---------------------

If you find any bug that may be a security problem, please follow our instructions [in our security policy](https://github.com/SAP-cloud-infrastructure/.github/blob/main/SECURITY.md) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

Code of Conduct
---------------

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/SAP-cloud-infrastructure/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

Licensing
---------

Copyright 2018-2025 SAP SE or an SAP affiliate company and terraform-provider-sci contributors. This repository contains code from [terraform-provider-openstack](https://github.com/terraform-provider-openstack/terraform-provider-openstack), copyright 2017-2025 terraform-provider-openstack contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/SAP-cloud-infrastructure/terraform-provider-sci).
