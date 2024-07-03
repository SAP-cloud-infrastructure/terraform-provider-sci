Converged Cloud - Terraform Provider
=======================================================

Documentation: [registry.terraform.io](https://registry.terraform.io/providers/sapcc/ccloud/latest/docs)

Maintainers
-----------

This provider plugin is maintained by:

  * [@kayrus](https://github.com/kayrus)
  * [@bugroger](https://github.com/BugRoger)

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 1.0.x
- [Go](https://golang.org/doc/install) 1.20 (to build the provider plugin)

Usage
---------------------

The CCloudEE provider is an extension to the [OpenStack Terraform
Provider](https://github.com/terraform-provider-openstack/terraform-provider-openstack).
It provides resources that allow to use Terraform for Converged Cloud's
additional services:

  * Kubernikus (Kubernetes as a Service)
  * Arc for Arc resources management
  * Lyra for Automation management
  * Billing for Billing management

The provider needs to be configured with the proper OpenStack credentials
before it can be used. For details see the OpenStack provider.

Building The Provider
---------------------

Clone the repository

```sh
$ git clone git@github.com:sapcc/terraform-provider-ccloud
```

Enter the provider directory and build the provider

```sh
$ cd terraform-provider-ccloud
$ make build
```

Installing the provider
-----------------------

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```hcl
terraform {
  required_providers {
    ccloud = {
      source = "sapcc/ccloud"
    }
  }
}

provider "ccloud" {
  # Configuration options
}
```

Using the provider
----------------------
Please see the documentation at [registry.terraform.io](https://registry.terraform.io/providers/sapcc/ccloud/latest/docs).

Or you can browse the documentation within this repo [here](https://github.com/sapcc/terraform-provider-ccloud/tree/master/website/docs).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](https://golang.org) installed on your machine (version 1.20+ is *required*).

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
build artifacts matching the [layout required](https://www.terraform.io/docs/registry/providers/publishing.html#manually-preparing-a-release)
to publish the provider in the Terraform Registry.

Releases will as drafts. Once marked as published on the GitHub Releases page,
they will become available via the Terraform Registry.
