package sci

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-provider-openstack/utils/v2/auth"
	"github.com/terraform-provider-openstack/utils/v2/mutexkv"
)

var version = "dev"

// Use openstackbase.Config as the base/foundation of this provider's
// Config struct.
type Config struct {
	auth.Config
}

// Provider returns a schema.Provider for OpenStack.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_AUTH_URL", ""),
				Description: descriptions["auth_url"],
			},

			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["region"],
				DefaultFunc: schema.EnvDefaultFunc("OS_REGION_NAME", ""),
			},

			"user_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USERNAME", ""),
				Description: descriptions["user_name"],
			},

			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USER_ID", ""),
				Description: descriptions["user_id"],
			},

			"application_credential_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_APPLICATION_CREDENTIAL_ID", ""),
				Description: descriptions["application_credential_id"],
			},

			"application_credential_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_APPLICATION_CREDENTIAL_NAME", ""),
				Description: descriptions["application_credential_name"],
			},

			"application_credential_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_APPLICATION_CREDENTIAL_SECRET", ""),
				Description: descriptions["application_credential_secret"],
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"OS_TENANT_ID",
					"OS_PROJECT_ID",
				}, ""),
				Description: descriptions["tenant_id"],
			},

			"tenant_name": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"OS_TENANT_NAME",
					"OS_PROJECT_NAME",
				}, ""),
				Description: descriptions["tenant_name"],
			},

			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("OS_PASSWORD", ""),
				Description: descriptions["password"],
			},

			"token": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"OS_TOKEN",
					"OS_AUTH_TOKEN",
				}, ""),
				Description: descriptions["token"],
			},

			"user_domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USER_DOMAIN_NAME", ""),
				Description: descriptions["user_domain_name"],
			},

			"user_domain_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USER_DOMAIN_ID", ""),
				Description: descriptions["user_domain_id"],
			},

			"project_domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_PROJECT_DOMAIN_NAME", ""),
				Description: descriptions["project_domain_name"],
			},

			"project_domain_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_PROJECT_DOMAIN_ID", ""),
				Description: descriptions["project_domain_id"],
			},

			"domain_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DOMAIN_ID", ""),
				Description: descriptions["domain_id"],
			},

			"domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DOMAIN_NAME", ""),
				Description: descriptions["domain_name"],
			},

			"default_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DEFAULT_DOMAIN", "default"),
				Description: descriptions["default_domain"],
			},

			"system_scope": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_SYSTEM_SCOPE", false),
				Description: descriptions["system_scope"],
			},

			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_INSECURE", nil),
				Description: descriptions["insecure"],
			},

			"endpoint_type": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_ENDPOINT_TYPE", ""),
			},

			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_CACERT", ""),
				Description: descriptions["cacert_file"],
			},

			"cert": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_CERT", ""),
				Description: descriptions["cert"],
			},

			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_KEY", ""),
				Description: descriptions["key"],
			},

			"delayed_auth": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_DELAYED_AUTH", true),
				Description: descriptions["delayed_auth"],
			},

			"allow_reauth": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_ALLOW_REAUTH", true),
				Description: descriptions["allow_reauth"],
			},

			"cloud": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_CLOUD", ""),
				Description: descriptions["cloud"],
			},

			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: descriptions["max_retries"],
			},

			"endpoint_overrides": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: descriptions["endpoint_overrides"],
			},

			"disable_no_cache_header": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["disable_no_cache_header"],
			},

			"enable_logging": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["enable_logging"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"sci_arc_agent_v1":               dataSourceSCIArcAgentV1(),
			"sci_arc_agent_ids_v1":           dataSourceSCIArcAgentIDsV1(),
			"sci_arc_job_v1":                 dataSourceSCIArcJobV1(),
			"sci_arc_job_ids_v1":             dataSourceSCIArcJobIDsV1(),
			"sci_automation_v1":              dataSourceSCIAutomationV1(),
			"sci_billing_domain_masterdata":  dataSourceSCIBillingDomainMasterdata(),
			"sci_billing_project_masterdata": dataSourceSCIBillingProjectMasterdata(),
			"sci_gslb_services_v1":           dataSourceSCIGSLBServicesV1(),
			"sci_endpoint_service_v1":        dataSourceSCIEndpointServiceV1(),
			"sci_networking_router_v2":       dataSourceSCINetworkingRouterV2(),
			// old provider names
			"ccloud_arc_agent_v1":               dataSourceSCIArcAgentV1(),
			"ccloud_arc_agent_ids_v1":           dataSourceSCIArcAgentIDsV1(),
			"ccloud_arc_job_v1":                 dataSourceSCIArcJobV1(),
			"ccloud_arc_job_ids_v1":             dataSourceSCIArcJobIDsV1(),
			"ccloud_automation_v1":              dataSourceSCIAutomationV1(),
			"ccloud_billing_domain_masterdata":  dataSourceSCIBillingDomainMasterdata(),
			"ccloud_billing_project_masterdata": dataSourceSCIBillingProjectMasterdata(),
			"ccloud_gslb_services_v1":           dataSourceSCIGSLBServicesV1(),
			"ccloud_endpoint_service_v1":        dataSourceSCIEndpointServiceV1(),
			"ccloud_networking_router_v2":       dataSourceSCINetworkingRouterV2(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"sci_arc_agent_bootstrap_v1":     resourceSCIArcAgentBootstrapV1(),
			"sci_arc_agent_v1":               resourceSCIArcAgentV1(),
			"sci_arc_job_v1":                 resourceSCIArcJobV1(),
			"sci_automation_v1":              resourceSCIAutomationV1(),
			"sci_automation_run_v1":          resourceSCIAutomationRunV1(),
			"sci_billing_domain_masterdata":  resourceSCIBillingDomainMasterdata(),
			"sci_billing_project_masterdata": resourceSCIBillingProjectMasterdata(),
			"sci_kubernetes_v1":              resourceSCIKubernetesV1(),
			"sci_bgpvpn_interconnection_v2":  resourceSCIBGPVPNInterconnectionV2(),
			"sci_gslb_datacenter_v1":         resourceSCIGSLBDatacenterV1(),
			"sci_gslb_domain_v1":             resourceSCIGSLBDomainV1(),
			"sci_gslb_pool_v1":               resourceSCIGSLBPoolV1(),
			"sci_gslb_member_v1":             resourceSCIGSLBMemberV1(),
			"sci_gslb_monitor_v1":            resourceSCIGSLBMonitorV1(),
			"sci_gslb_quota_v1":              resourceSCIGSLBQuotaV1(),
			"sci_gslb_geomap_v1":             resourceSCIGSLBGeoMapV1(),
			"sci_endpoint_service_v1":        resourceSCIEndpointServiceV1(),
			"sci_endpoint_v1":                resourceSCIEndpointV1(),
			"sci_endpoint_accept_v1":         resourceSCIEndpointAcceptV1(),
			"sci_endpoint_quota_v1":          resourceSCIEndpointQuotaV1(),
			"sci_endpoint_rbac_policy_v1":    resourceSCIEndpointRBACV1(),
			// old provider names
			"ccloud_arc_agent_bootstrap_v1":     resourceSCIArcAgentBootstrapV1(),
			"ccloud_arc_agent_v1":               resourceSCIArcAgentV1(),
			"ccloud_arc_job_v1":                 resourceSCIArcJobV1(),
			"ccloud_automation_v1":              resourceSCIAutomationV1(),
			"ccloud_automation_run_v1":          resourceSCIAutomationRunV1(),
			"ccloud_billing_domain_masterdata":  resourceSCIBillingDomainMasterdata(),
			"ccloud_billing_project_masterdata": resourceSCIBillingProjectMasterdata(),
			"ccloud_kubernetes_v1":              resourceSCIKubernetesV1(),
			"ccloud_kubernetes":                 resourceSCIKubernetesV1(),
			"ccloud_bgpvpn_interconnection_v2":  resourceSCIBGPVPNInterconnectionV2(),
			"ccloud_gslb_datacenter_v1":         resourceSCIGSLBDatacenterV1(),
			"ccloud_gslb_domain_v1":             resourceSCIGSLBDomainV1(),
			"ccloud_gslb_pool_v1":               resourceSCIGSLBPoolV1(),
			"ccloud_gslb_member_v1":             resourceSCIGSLBMemberV1(),
			"ccloud_gslb_monitor_v1":            resourceSCIGSLBMonitorV1(),
			"ccloud_gslb_quota_v1":              resourceSCIGSLBQuotaV1(),
			"ccloud_gslb_geomap_v1":             resourceSCIGSLBGeoMapV1(),
			"ccloud_endpoint_service_v1":        resourceSCIEndpointServiceV1(),
			"ccloud_endpoint_v1":                resourceSCIEndpointV1(),
			"ccloud_endpoint_accept_v1":         resourceSCIEndpointAcceptV1(),
			"ccloud_endpoint_quota_v1":          resourceSCIEndpointQuotaV1(),
			"ccloud_endpoint_rbac_policy_v1":    resourceSCIEndpointRBACV1(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return configureProvider(ctx, d, terraformVersion)
	}

	return provider
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"auth_url": "The Identity authentication URL.",

		"cloud": "An entry in a `clouds.yaml` file to use.",

		"region": "The OpenStack region to connect to.",

		"user_name": "Username to login with.",

		"user_id": "User ID to login with.",

		"application_credential_id": "Application Credential ID to login with.",

		"application_credential_name": "Application Credential name to login with.",

		"application_credential_secret": "Application Credential secret to login with.",

		"tenant_id": "The ID of the Tenant (Identity v2) or Project (Identity v3)\n" +
			"to login with.",

		"tenant_name": "The name of the Tenant (Identity v2) or Project (Identity v3)\n" +
			"to login with.",

		"password": "Password to login with.",

		"token": "Authentication token to use as an alternative to username/password.",

		"user_domain_name": "The name of the domain where the user resides (Identity v3).",

		"user_domain_id": "The ID of the domain where the user resides (Identity v3).",

		"project_domain_name": "The name of the domain where the project resides (Identity v3).",

		"project_domain_id": "The ID of the domain where the proejct resides (Identity v3).",

		"domain_id": "The ID of the Domain to scope to (Identity v3).",

		"domain_name": "The name of the Domain to scope to (Identity v3).",

		"default_domain": "The name of the Domain ID to scope to if no other domain is specified. Defaults to `default` (Identity v3).",

		"system_scope": "If set to `true`, system scoped authorization will be enabled. Defaults to `false` (Identity v3).",

		"insecure": "Trust self-signed certificates.",

		"cacert_file": "A Custom CA certificate.",

		"cert": "A client certificate to authenticate with.",

		"key": "A client private key to authenticate with.",

		"endpoint_type": "The catalog endpoint type to use.",

		"endpoint_overrides": "A map of services with an endpoint to override what was\n" +
			"from the Keystone catalog",

		"disable_no_cache_header": "If set to `true`, the HTTP `Cache-Control: no-cache` header will not be added by default to all API requests.",

		"delayed_auth": "If set to `false`, OpenStack authorization will be perfomed,\n" +
			"every time the service provider client is called. Defaults to `true`.",

		"allow_reauth": "If set to `false`, OpenStack authorization won't be perfomed\n" +
			"automatically, if the initial auth token get expired. Defaults to `true`",

		"max_retries": "How many times HTTP connection should be retried until giving up.",

		"enable_logging": "Outputs very verbose logs with all calls made to and responses from OpenStack",
	}
}

func getSDKVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, v := range buildInfo.Deps {
		if v.Path == "github.com/hashicorp/terraform-plugin-sdk/v2" {
			return v.Version
		}
	}

	return ""
}

func configureProvider(ctx context.Context, d *schema.ResourceData, terraformVersion string) (any, diag.Diagnostics) {
	enableLogging := d.Get("enable_logging").(bool)
	if !enableLogging {
		// enforce logging (similar to OS_DEBUG) when TF_LOG is 'DEBUG' or 'TRACE'
		if logLevel := logging.LogLevel(); logLevel != "" && os.Getenv("OS_DEBUG") == "" {
			if logLevel == "DEBUG" || logLevel == "TRACE" {
				enableLogging = true
			}
		}
	}

	authOpts := &gophercloud.AuthOptions{
		Scope: &gophercloud.AuthScope{System: d.Get("system_scope").(bool)},
	}

	config := Config{
		auth.Config{
			CACertFile:                  d.Get("cacert_file").(string),
			ClientCertFile:              d.Get("cert").(string),
			ClientKeyFile:               d.Get("key").(string),
			Cloud:                       d.Get("cloud").(string),
			DefaultDomain:               d.Get("default_domain").(string),
			DomainID:                    d.Get("domain_id").(string),
			DomainName:                  d.Get("domain_name").(string),
			EndpointOverrides:           d.Get("endpoint_overrides").(map[string]any),
			EndpointType:                d.Get("endpoint_type").(string),
			IdentityEndpoint:            d.Get("auth_url").(string),
			Password:                    d.Get("password").(string),
			ProjectDomainID:             d.Get("project_domain_id").(string),
			ProjectDomainName:           d.Get("project_domain_name").(string),
			Region:                      d.Get("region").(string),
			Token:                       d.Get("token").(string),
			TenantID:                    d.Get("tenant_id").(string),
			TenantName:                  d.Get("tenant_name").(string),
			UserDomainID:                d.Get("user_domain_id").(string),
			UserDomainName:              d.Get("user_domain_name").(string),
			Username:                    d.Get("user_name").(string),
			UserID:                      d.Get("user_id").(string),
			ApplicationCredentialID:     d.Get("application_credential_id").(string),
			ApplicationCredentialName:   d.Get("application_credential_name").(string),
			ApplicationCredentialSecret: d.Get("application_credential_secret").(string),
			DelayedAuth:                 d.Get("delayed_auth").(bool),
			AllowReauth:                 d.Get("allow_reauth").(bool),
			AuthOpts:                    authOpts,
			MaxRetries:                  d.Get("max_retries").(int),
			DisableNoCacheHeader:        d.Get("disable_no_cache_header").(bool),
			TerraformVersion:            terraformVersion,
			SDKVersion:                  getSDKVersion() + " Terraform Provider SCI/" + version,
			MutexKV:                     mutexkv.NewMutexKV(),
			EnableLogger:                enableLogging,
		},
	}

	v, ok := getOkExists(d, "insecure")
	if ok {
		insecure := v.(bool)
		config.Insecure = &insecure
	}

	if err := config.LoadAndValidate(ctx); err != nil {
		return nil, diag.FromErr(err)
	}

	return &config, nil
}
