package sci

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sapcc/kubernikus/pkg/api/client/operations"
	"github.com/sapcc/kubernikus/pkg/api/models"
)

func resourceSCIKubernetesV1() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,

		ReadContext:   resourceSCIKubernetesV1Read,
		UpdateContext: resourceSCIKubernetesV1Update,
		CreateContext: resourceSCIKubernetesV1Create,
		DeleteContext: resourceSCIKubernetesV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSCIKubernetesV1Import,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: kubernikusValidateClusterName,
			},

			"is_admin": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},

			"advertise_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPAddress,
			},

			"advertise_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 65536),
			},

			"audit": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"elasticsearch",
					"swift",
					"http",
					"stdout",
				}, false),
			},

			"cluster_cidr": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "100.100.0.0/16",
				ValidateFunc: func(v any, k string) ([]string, []error) {
					if v == nil || v.(string) == "" {
						return nil, nil
					}
					return validation.IsCIDRNetwork(8, 17)(v, k)
				},
			},

			"service_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IsCIDRNetwork(8, 24),
			},

			"dns_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPAddress,
			},

			"dns_domain": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"ssh_public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"no_cloud": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},

			"dex": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  true,
			},

			"oidc": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"issuer_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"client_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"authentication_configuration": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: kubernikusValidateAuthConf,
			},

			"dashboard": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  true,
			},

			"backup": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					models.KlusterSpecBackupOn, models.KlusterSpecBackupOff, models.KlusterSpecBackupExternalAWS,
				}, false),
			},

			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateKubernetesVersion,
			},

			"node_pools": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: kubernikusValidatePoolName,
						},
						"flavor": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"image": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 127),
						},
						"availability_zone": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"taints": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"labels": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"custom_root_disk_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(64, 1024),
						},
						"config": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow_reboot": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"allow_replace": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"openstack": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lb_floating_network_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"network_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"lb_subnet_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"router_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"security_group_name": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},

			"phase": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"wormhole": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"apiserver_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dashboard_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"kube_config": {
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_key": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"cluster_ca_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"not_before": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"not_after": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"kube_config_raw": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceSCIKubernetesV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Creating Kubernikus Kluster in project %s", config.TenantID)

	klient, err := config.kubernikusV1Client(ctx, GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return diag.Errorf("Error creating Kubernikus client: %s", err)
	}

	cluster := &models.Kluster{
		Spec: models.KlusterSpec{
			NodePools: []models.NodePool{},
			Openstack: models.OpenstackSpec{},
		},
	}

	cluster.Name = d.Get("name").(string)
	cluster.Spec.AdvertiseAddress = d.Get("advertise_address").(string)
	cluster.Spec.AdvertisePort = int64(d.Get("advertise_port").(int))
	if v, ok := d.GetOk("audit"); ok {
		auditStr := v.(string)
		cluster.Spec.Audit = &auditStr
	}
	if v, ok := d.Get("cluster_cidr").(string); ok {
		cluster.Spec.ClusterCIDR = &v
	}
	cluster.Spec.DNSAddress = d.Get("dns_address").(string)
	cluster.Spec.DNSDomain = d.Get("dns_domain").(string)
	cluster.Spec.SSHPublicKey = d.Get("ssh_public_key").(string)
	cluster.Spec.NoCloud = d.Get("no_cloud").(bool)
	if v, ok := d.Get("dex").(bool); ok {
		cluster.Spec.Dex = &v
	}
	if v, ok := d.GetOk("oidc"); ok {
		cluster.Spec.Oidc = kubernikusExpandOIDCV1(v.([]any))
	}
	if v, ok := d.Get("authentication_configuration").(string); ok && v != "" {
		cluster.Spec.AuthenticationConfiguration = models.AuthenticationConfiguration(v)
	}
	if v, ok := d.Get("dashboard").(bool); ok {
		cluster.Spec.Dashboard = &v
	}
	cluster.Spec.Backup = d.Get("backup").(string)
	cluster.Spec.ServiceCIDR = d.Get("service_cidr").(string)
	if v, ok := d.GetOk("version"); ok {
		v := v.(string)
		err = verifySupportedKubernetesVersion(klient, v)
		if err != nil {
			return diag.FromErr(err)
		}
		cluster.Spec.Version = v
	}
	cluster.Spec.NodePools, err = kubernikusExpandNodePoolsV1(d.Get("node_pools"))
	if err != nil {
		return diag.FromErr(err)
	}
	if v := kubernikusExpandOpenstackSpecV1(d.Get("openstack")); v != nil {
		cluster.Spec.Openstack = *v
	}

	_, err = klient.CreateCluster(operations.NewCreateClusterParams().WithBody(cluster), klient.authFunc())
	if err != nil {
		return diag.FromErr(kubernikusHandleErrorV1("Error creating cluster", err))
	}

	d.SetId(cluster.Name)

	// waiting for Running state
	timeout := d.Timeout(schema.TimeoutCreate)
	target := string(models.KlusterPhaseRunning)
	pending := []string{
		string(models.KlusterPhasePending),
		string(models.KlusterPhaseCreating),
		string(models.KlusterPhaseUpgrading),
	}
	err = kubernikusWaitForClusterV1(ctx, klient, cluster.Name, target, pending, timeout)
	if err != nil {
		return diag.FromErr(kubernikusHandleErrorV1("Error waiting for running cluster state", err))
	}

	return resourceSCIKubernetesV1Read(ctx, d, meta)
}

func resourceSCIKubernetesV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Reading Kubernikus Kluster in project %s", config.TenantID)

	klient, err := config.kubernikusV1Client(ctx, GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return diag.Errorf("Error creating Kubernikus client: %s", err)
	}

	result, err := klient.ShowCluster(operations.NewShowClusterParams().WithName(d.Id()), klient.authFunc())
	if err != nil {
		switch res := err.(type) {
		case *operations.ShowClusterDefault:
			if res.Payload.Message == "Not found" {
				d.SetId("")
				return nil
			}

			return diag.Errorf("Error reading Kubernikus cluster: %s", res.Payload.Message)
		case error:
			return diag.Errorf("Error reading Kubernikus cluster: %s", err)
		}
		return diag.FromErr(err)
	}

	_ = d.Set("advertise_address", result.Payload.Spec.AdvertiseAddress)
	_ = d.Set("advertise_port", result.Payload.Spec.AdvertisePort)
	_ = d.Set("audit", result.Payload.Spec.Audit)
	_ = d.Set("cluster_cidr", result.Payload.Spec.ClusterCIDR)
	_ = d.Set("dns_address", result.Payload.Spec.DNSAddress)
	_ = d.Set("dns_domain", result.Payload.Spec.DNSDomain)
	_ = d.Set("name", result.Payload.Spec.Name)
	_ = d.Set("ssh_public_key", result.Payload.Spec.SSHPublicKey)
	_ = d.Set("no_cloud", result.Payload.Spec.NoCloud)
	_ = d.Set("dex", result.Payload.Spec.Dex)
	_ = d.Set("oidc", kubernikusFlattenOIDCV1(result.Payload.Spec.Oidc))
	_ = d.Set("authentication_configuration", result.Payload.Spec.AuthenticationConfiguration)
	_ = d.Set("dashboard", result.Payload.Spec.Dashboard)
	_ = d.Set("backup", result.Payload.Spec.Backup)
	_ = d.Set("service_cidr", result.Payload.Spec.ServiceCIDR)
	_ = d.Set("version", result.Payload.Spec.Version)
	_ = d.Set("phase", result.Payload.Status.Phase)
	_ = d.Set("wormhole", result.Payload.Status.Wormhole)
	_ = d.Set("apiserver_url", result.Payload.Status.Apiserver)
	_ = d.Set("dashboard_url", result.Payload.Status.Dashboard)
	_ = d.Set("openstack", kubernikusFlattenOpenstackSpecV1(&result.Payload.Spec.Openstack))
	_ = d.Set("node_pools", kubernikusFlattenNodePoolsV1(result.Payload.Spec.NodePools))

	_ = d.Set("region", GetRegion(d, config))

	// if cluster is in pending state, than there are no credentials yet
	if result.Payload.Status.Phase != models.KlusterPhasePending {
		kubeConfigRaw, kubeConfig, err := getCredentials(klient, d.Id(), d.Get("kube_config_raw").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		_ = d.Set("kube_config", kubeConfig)
		_ = d.Set("kube_config_raw", kubeConfigRaw)
	}

	return nil
}

func resourceSCIKubernetesV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Updating Kubernikus Kluster in project %s", config.TenantID)

	klient, err := config.kubernikusV1Client(ctx, GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return diag.Errorf("Error creating Kubernikus client: %s", err)
	}

	timeout := d.Timeout(schema.TimeoutUpdate)
	cluster := &models.Kluster{
		Spec: models.KlusterSpec{
			NodePools: []models.NodePool{},
			Openstack: models.OpenstackSpec{},
		},
	}

	cluster.Name = d.Id()

	if v, ok := d.GetOk("audit"); ok {
		auditStr := v.(string)
		cluster.Spec.Audit = &auditStr
	}

	if v, ok := d.GetOk("ssh_public_key"); ok {
		cluster.Spec.SSHPublicKey = v.(string)
	}

	if v, ok := d.GetOk("backup"); ok {
		cluster.Spec.Backup = v.(string)
	}

	if v, ok := d.Get("dex").(bool); ok {
		cluster.Spec.Dex = &v
	}

	if v, ok := d.GetOk("oidc"); ok {
		cluster.Spec.Oidc = kubernikusExpandOIDCV1(v.([]any))
	}

	if v, ok := d.Get("authentication_configuration").(string); ok && v != "" {
		cluster.Spec.AuthenticationConfiguration = models.AuthenticationConfiguration(v)
	}

	if v, ok := d.Get("dashboard").(bool); ok {
		cluster.Spec.Dashboard = &v
	}

	if v, ok := d.GetOk("version"); ok {
		cluster.Spec.Version = v.(string)
		err = verifySupportedKubernetesVersion(klient, cluster.Spec.Version)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("openstack.0.security_group_name"); ok {
		cluster.Spec.Openstack.SecurityGroupName = v.(string)
	}

	o, n := d.GetChange("node_pools")

	// wait for the cluster to be upgraded, when new API version was specified
	target := string(models.KlusterPhaseRunning)
	pending := []string{
		string(models.KlusterPhasePending),
		string(models.KlusterPhaseCreating),
		string(models.KlusterPhaseUpgrading),
		string(models.KlusterPhaseTerminating),
	}
	err = kubernikusUpdateNodePoolsV1(ctx, klient, cluster, o, n, target, pending, timeout)
	if err != nil {
		return diag.FromErr(kubernikusHandleErrorV1("Error waiting for cluster to be updated", err))
	}

	return resourceSCIKubernetesV1Read(ctx, d, meta)
}

func resourceSCIKubernetesV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Deleting Kubernikus Kluster in project %s", config.TenantID)

	klient, err := config.kubernikusV1Client(ctx, GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return diag.Errorf("Error creating Kubernikus client: %s", err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)

	_, err = klient.TerminateCluster(operations.NewTerminateClusterParams().WithName(d.Id()), klient.authFunc())
	if err != nil {
		return diag.FromErr(kubernikusHandleErrorV1("Error deleting cluster", err))
	}

	target := "Terminated"
	pending := []string{
		string(models.KlusterPhasePending),
		string(models.KlusterPhaseCreating),
		string(models.KlusterPhaseRunning),
		string(models.KlusterPhaseUpgrading),
		string(models.KlusterPhaseTerminating),
	}
	err = kubernikusWaitForClusterV1(ctx, klient, d.Id(), target, pending, timeout)
	if err != nil {
		return diag.FromErr(kubernikusHandleErrorV1("Error waiting for cluster to be deleted", err))
	}

	return nil
}

func resourceSCIKubernetesV1Import(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	name := parts[0]

	if len(name) == 0 {
		err := fmt.Errorf("invalid format specified for Kubernetes, format must be <name>[/<is_admin>]")
		return nil, err
	}

	var isAdmin bool
	var err error
	if len(parts) == 2 {
		isAdmin, err = strconv.ParseBool(parts[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse is_admin field: %s", err)
		}
	}

	d.SetId(name)
	_ = d.Set("is_admin", isAdmin)

	return []*schema.ResourceData{d}, nil
}
