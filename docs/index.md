---
layout: "sci"
page_title: "Provider: SAP Cloud Infrastructure"
sidebar_current: "docs-sci-index"
description: |-
  The SAP Cloud Infrastructure provider is used to interact with the many resources supported by SAP Cloud Infrastructure. The provider needs to be configured with the proper credentials before it can be used.
---

# SAP Cloud Infrastructure Provider

The SAP Cloud Infrastructure provider is used to interact with the
many resources supported by SAP Cloud Infrastructure. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

# CCloud Provider Backward Compatibility

The SAP Cloud Infrastructure provider is the successor to the original CCloud
provider. It is designed to be backward compatible with the CCloud provider, so
that existing Terraform configurations using the CCloud provider can be
migrated to the SAP Cloud Infrastructure provider with minimal changes.

To migrate from the CCloud provider to the SAP Cloud Infrastructure provider,
you can simply change the provider source in your Terraform configuration from
`ccloud` to `sci`. For example, change:

```hcl
terraform {
  required_providers {
    ccloud = {
      source = "sapcc/ccloud"
    }
  }
}
```

to:

```hcl
terraform {
  required_providers {
    ccloud = {
      source = "SAP-cloud-infrastructure/sci"
    }
  }
}
```

and update the state file to reflect the new provider source. You can do this by running the following command:

```shell
$ terraform state replace-provider sapcc/ccloud SAP-cloud-infrastructure/sci
```

# Installing the provider

## Example Usage

```hcl
# Define required providers
terraform {
  required_providers {
    sci = {
      source = "SAP-cloud-infrastructure/sci"
    }
  }
}

# Configure the SAP Cloud Infrastructure Provider
provider "sci" {
  user_name   = "admin"
  tenant_name = "admin"
  password    = "pwd"
  auth_url    = "http://myauthurl:5000/v2.0"
  region      = "RegionOne"
}

# Create a Kubernetes cluster
resource "sci_kubernetes_v1" "test-cluster" {
  # ...
}
```

## Configuration Reference

The following arguments are supported:

* `auth_url` - (Optional; required if `cloud` is not specified) The Identity
  authentication URL. If omitted, the `OS_AUTH_URL` environment variable is used.

* `cloud` - (Optional; required if `auth_url` is not specified) An entry in a
  `clouds.yaml` file. See the OpenStack `openstacksdk`
  [documentation](https://docs.openstack.org/openstacksdk/latest/user/config/configuration.html)
  for more information about `clouds.yaml` files. If omitted, the `OS_CLOUD`
  environment variable is used.

* `region` - (Optional) The region of the SAP Cloud Infrastructure to use. If omitted,
  the `OS_REGION_NAME` environment variable is used. If `OS_REGION_NAME` is
  not set, then no region will be used. It should be possible to omit the
  region in single-region SAP Cloud Infrastructure environments, but this behavior may vary
  depending on the SAP Cloud Infrastructure environment being used.

* `user_name` - (Optional) The Username to login with. If omitted, the
  `OS_USERNAME` environment variable is used.

* `user_id` - (Optional) The User ID to login with. If omitted, the
  `OS_USER_ID` environment variable is used.

* `application_credential_id` - (Optional) (Identity v3 only) The ID of an
  application credential to authenticate with. An
  `application_credential_secret` has to bet set along with this parameter.
  If omitted, the `OS_APPLICATION_CREDENTIAL_ID` environment variable is used.

* `application_credential_name` - (Optional) (Identity v3 only) The name of an
  application credential to authenticate with. Requires `user_id`, or
  `user_name` and `user_domain_name` (or `user_domain_id`) to be set.
  If omitted, the `OS_APPLICATION_CREDENTIAL_NAME` environment variable is used.

* `application_credential_secret` - (Optional) (Identity v3 only) The secret of
  an application credential to authenticate with. Required by
  `application_credential_id` or `application_credential_name`.
  If omitted, the `OS_APPLICATION_CREDENTIAL_SECRET` environment variable is used.

* `tenant_id` - (Optional) The ID of the Tenant (Identity v2) or Project
  (Identity v3) to login with. If omitted, the `OS_TENANT_ID` or
  `OS_PROJECT_ID` environment variables are used.

* `tenant_name` - (Optional) The Name of the Tenant (Identity v2) or Project
  (Identity v3) to login with. If omitted, the `OS_TENANT_NAME` or
  `OS_PROJECT_NAME` environment variable are used.

* `password` - (Optional) The Password to login with. If omitted, the
  `OS_PASSWORD` environment variable is used.

* `token` - (Optional; Required if not using `user_name` and `password`)
  A token is an expiring, temporary means of access issued via the Keystone
  service. By specifying a token, you do not have to specify a username/password
  combination, since the token was already created by a username/password out of
  band of Terraform. If omitted, the `OS_TOKEN` or `OS_AUTH_TOKEN` environment
  variables are used.

* `user_domain_name` - (Optional) The domain name where the user is located. If
  omitted, the `OS_USER_DOMAIN_NAME` environment variable is checked.

* `user_domain_id` - (Optional) The domain ID where the user is located. If
  omitted, the `OS_USER_DOMAIN_ID` environment variable is checked.

* `project_domain_name` - (Optional) The domain name where the project is
  located. If omitted, the `OS_PROJECT_DOMAIN_NAME` environment variable is
  checked.

* `project_domain_id` - (Optional) The domain ID where the project is located
  If omitted, the `OS_PROJECT_DOMAIN_ID` environment variable is checked.

* `domain_id` - (Optional) The ID of the Domain to scope to (Identity v3). If
  omitted, the `OS_DOMAIN_ID` environment variable is checked.

* `domain_name` - (Optional) The Name of the Domain to scope to (Identity v3).
  If omitted, the following environment variables are checked (in this order):
  `OS_DOMAIN_NAME`.

* `default_domain` - (Optional) The ID of the Domain to scope to if no other
  domain is specified (Identity v3). If omitted, the environment variable
  `OS_DEFAULT_DOMAIN` is checked or a default value of "default" will be
  used.

* `system_scope` - (Optional) Set to `true` to enable system scoped authorization. If omitted, the `OS_SYSTEM_SCOPE` environment variable is used.

* `insecure` - (Optional) Trust self-signed SSL certificates. If omitted, the
  `OS_INSECURE` environment variable is used.

* `cacert_file` - (Optional) Specify a custom CA certificate when communicating
  over SSL. You can specify either a path to the file or the contents of the
  certificate. If omitted, the `OS_CACERT` environment variable is used.

* `cert` - (Optional) Specify client certificate file for SSL client
  authentication. You can specify either a path to the file or the contents of
  the certificate. If omitted the `OS_CERT` environment variable is used.

* `key` - (Optional) Specify client private key file for SSL client
  authentication. You can specify either a path to the file or the contents of
  the key. If omitted the `OS_KEY` environment variable is used.

* `endpoint_type` - (Optional) Specify which type of endpoint to use from the
  service catalog. It can be set using the `OS_ENDPOINT_TYPE` environment
  variable. If not set, public endpoints is used.

* `endpoint_overrides` - (Optional) A set of key/value pairs that can
  override an endpoint for a specified SAP Cloud Infrastructure service. Setting an override
  requires you to specify the full and complete endpoint URL. This might
  also invalidate any region you have set, too. Please see below for more details.
  Please use this at your own risk.

* `disable_no_cache_header` - (Optional) If set to `true`, the HTTP
  `Cache-Control: no-cache` header will not be added by default to all API requests.
  If omitted this header is added to all API requests to force HTTP caches (if any)
  to go upstream instead of serving cached responses.

* `delayed_auth` - (Optional) If set to `false`, OpenStack authorization will be perfomed,
  every time the service provider client is called. Defaults to `true`.
  If omitted, the `OS_DELAYED_AUTH` environment variable is checked.

* `allow_reauth` - (Optional) If set to `false`, OpenStack authorization won't be
  perfomed automatically, if the initial auth token get expired. Defaults to `true`.
  If omitted, the `OS_ALLOW_REAUTH` environment variable is checked.

* `max_retries` - (Optional) If set to a value greater than 0, the OpenStack
  client will retry failed HTTP connections and Too Many Requests (429 code)
  HTTP responses with a `Retry-After` header within the specified value.

* `enable_logging` - (Optional) When enabled, generates verbose logs containing
  all the calls made to and responses received from OpenStack.

## Overriding Service API Endpoints

There might be a situation in which you want or need to override an API endpoint
rather than use the endpoint which was returned to you in the service catalog.
You can do this by configuring the `endpoint_overrides` argument in the provider
configuration:

```hcl
provider "sci" {

  endpoint_overrides = {
    "arc"               = "https://arc.example.com/api/v1/"
    "automation"        = "https://lyra.example.com:8776/api/v1/"
    "gtm"               = "https://gtm.example.com/v1"
    "endpoint-services" = "https://archer.example.com/v1"
  }

}
```

Note how each URL ends in a "/" and the `volumev2` service includes the
tenant/project UUID. You must make sure you specify the full and complete
endpoint URL for this to work.

The service keys are the standard service entries used in the SAP Cloud Infrastructure
Identity/Keystone service catalog. This provider supports:

* `arc`: Arc / Arc v1
* `automation`: Automation / Lyra v1
* `kubernikus`: Kubernetes / Kubernikus
* `sapcc-billing`: Billing
* `gtm`: Andromeda a GSLB / GTM (Global Server Load Balancing / Global Traffic Management) service
* `endpoint-services`: Archer / Endpoint Services

Please use this feature at your own risk. If you are unsure about needing
to override an endpoint, you most likely do not need to override one.

## Additional Logging

This provider has the ability to log all HTTP requests and responses between
Terraform and the SAP Cloud Infrastructure which is useful for troubleshooting and
debugging.

To enable these logs, set the `OS_DEBUG` environment variable to `1` along
with the usual `TF_LOG=DEBUG` environment variable:

```shell
$ OS_DEBUG=1 TF_LOG=DEBUG terraform apply
```

If you submit these logs with a bug report, please ensure any sensitive
information has been scrubbed first!
