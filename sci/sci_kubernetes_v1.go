package sci

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/sapcc/kubernikus/pkg/api/client/operations"
	"github.com/sapcc/kubernikus/pkg/api/models"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

const (
	klusterNameRegex = "^[a-z][-a-z0-9]{0,18}[a-z0-9]?$"
	poolNameRegex    = "^[a-z][-\\.a-z0-9]{0,18}[a-z0-9]?$"
)

func kubernikusValidateClusterName(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	if !regexp.MustCompile(klusterNameRegex).MatchString(value) {
		errors = append(errors,
			fmt.Errorf("%q must be 1 to 20 characters with lowercase and uppercase letters, numbers and hyphens", k))
	}
	return
}

func kubernikusValidatePoolName(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	if !regexp.MustCompile(poolNameRegex).MatchString(value) {
		errors = append(errors,
			fmt.Errorf("%q must be 1 to 20 characters with lowercase and uppercase letters, numbers, hyphens and dots", k))
	}
	return
}

func kubernikusValidateAuthConf(v any, k string) ([]string, []error) {
	if v == nil {
		return nil, nil
	}
	authConf, ok := v.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected string for %s, got %T", k, v)}
	}

	err := models.AuthenticationConfiguration(authConf).Validate(strfmt.Default)
	if err != nil {
		return nil, []error{fmt.Errorf("invalid authentication_configuration: %s", err)}
	}

	return nil, nil
}

func kubernikusFlattenOpenstackSpecV1(spec *models.OpenstackSpec) []map[string]any {
	var res []map[string]any

	if spec == (&models.OpenstackSpec{}) {
		return res
	}

	return append(res, map[string]any{
		"lb_floating_network_id": spec.LBFloatingNetworkID,
		"lb_subnet_id":           spec.LBSubnetID,
		"network_id":             spec.NetworkID,
		"router_id":              spec.RouterID,
		"security_group_name":    spec.SecurityGroupName,
	})
}

func kubernikusFlattenNodePoolsV1(nodePools []models.NodePool) []map[string]any {
	res := make([]map[string]any, 0, len(nodePools))
	for _, p := range nodePools {
		res = append(res, map[string]any{
			"availability_zone":     p.AvailabilityZone,
			"flavor":                p.Flavor,
			"image":                 p.Image,
			"name":                  p.Name,
			"size":                  p.Size,
			"taints":                p.Taints,
			"labels":                p.Labels,
			"custom_root_disk_size": p.CustomRootDiskSize,
			"config": []map[string]any{
				{
					"allow_reboot":  p.Config.AllowReboot,
					"allow_replace": p.Config.AllowReplace,
				},
			},
		})
	}
	return res
}

func kubernikusExpandOpenstackSpecV1(raw any) *models.OpenstackSpec {
	if raw != nil {
		if v, ok := raw.([]any); ok {
			for _, v := range v {
				if v, ok := v.(map[string]any); ok {
					res := new(models.OpenstackSpec)

					if v, ok := v["lb_floating_network_id"]; ok {
						res.LBFloatingNetworkID = v.(string)
					}
					if v, ok := v["lb_subnet_id"]; ok {
						res.LBSubnetID = v.(string)
					}
					if v, ok := v["network_id"]; ok {
						res.NetworkID = v.(string)
					}
					if v, ok := v["router_id"]; ok {
						res.RouterID = v.(string)
					}
					if v, ok := v["security_group_name"]; ok {
						res.SecurityGroupName = v.(string)
					}

					return res
				}
			}
		}
	}

	return nil
}

func kubernikusExpandNodePoolsV1(raw any) ([]models.NodePool, error) {
	var names []string

	if raw != nil {
		if v, ok := raw.([]any); ok {
			var res []models.NodePool

			for _, v := range v {
				if v, ok := v.(map[string]any); ok {
					var p models.NodePool

					if v, ok := v["name"]; ok {
						p.Name = v.(string)
						if strSliceContains(names, p.Name) {
							return nil, fmt.Errorf("duplicate node pool name found: %s", p.Name)
						}
						names = append(names, p.Name)
					}
					if v, ok := v["flavor"]; ok {
						p.Flavor = v.(string)
					}
					if v, ok := v["image"]; ok {
						p.Image = v.(string)
					}
					if v, ok := v["size"]; ok {
						p.Size = int64(v.(int))
					}
					if v, ok := v["availability_zone"]; ok {
						p.AvailabilityZone = v.(string)
					}
					if v, ok := v["taints"]; ok {
						p.Taints = expandToStringSlice(v.([]any))
					}
					if v, ok := v["labels"]; ok {
						p.Labels = expandToStringSlice(v.([]any))
					}
					if v, ok := v["custom_root_disk_size"]; ok {
						p.CustomRootDiskSize = int64(v.(int))
					}
					if v, ok := v["config"]; ok {
						p.Config = expandToNodePoolConfig(v.([]any))
					}

					res = append(res, p)
				}
			}

			return res, nil
		}
	}

	return nil, nil
}

func kubernikusWaitForClusterV1(ctx context.Context, klient *kubernikus, name string, target string, pending []string, timeout time.Duration) error {
	// Phase: "Pending","Creating","Running","Terminating","Upgrading"
	log.Printf("[DEBUG] Waiting for %s cluster to become %s.", name, target)

	stateConf := &retry.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    kubernikusKlusterV1GetPhase(klient, target, name),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if e, ok := err.(*operations.ShowClusterDefault); ok && target == "Terminated" && e.Payload.Message == "Not found" {
			return nil
		}
	}

	return err
}

func kubernikusKlusterV1GetPhase(klient *kubernikus, target string, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		result, err := klient.ShowCluster(operations.NewShowClusterParams().WithName(name), klient.authFunc())
		if err != nil {
			return nil, "", err
		}

		if target != "Terminated" {
			events, err := klient.GetClusterEvents(operations.NewGetClusterEventsParams().WithName(name), klient.authFunc())
			if err != nil {
				return nil, "", err
			}

			if len(events.Payload) > 0 {
				// check, whether there are error events
				event := events.Payload[len(events.Payload)-1]

				if strings.Contains(event.Reason, "Error") || strings.Contains(event.Reason, "Failed") {
					return nil, event.Reason, fmt.Errorf("%s", event.Message)
				}
			}

			for _, a := range result.Payload.Spec.NodePools {
				// workaround for the upgrade status race condition
				if result.Payload.Status.Phase == models.KlusterPhaseRunning &&
					result.Payload.Spec.Version != result.Payload.Status.ApiserverVersion {
					return result.Payload, string(models.KlusterPhaseUpgrading), nil
				}

				for _, s := range result.Payload.Status.NodePools {
					if a.Name == s.Name {
						// sometimes status size doesn't reflect the actual size, therefore we use "a.Size"
						if a.Size != s.Healthy {
							return result.Payload, "Pending", nil
						}
					}
				}
			}

			if len(result.Payload.Spec.NodePools) != len(result.Payload.Status.NodePools) {
				return result.Payload, "Pending", nil
			}
		}

		return result.Payload, string(result.Payload.Status.Phase), nil
	}
}

func kubernikusUpdateNodePoolsV1(ctx context.Context, klient *kubernikus, cluster *models.Kluster, oldNodePoolsRaw, newNodePoolsRaw any, target string, pending []string, timeout time.Duration) error {
	var poolsToKeep []models.NodePool
	var poolsToDelete []models.NodePool
	oldNodePools, err := kubernikusExpandNodePoolsV1(oldNodePoolsRaw)
	if err != nil {
		return err
	}
	newNodePools, err := kubernikusExpandNodePoolsV1(newNodePoolsRaw)
	if err != nil {
		return err
	}

	pretty, err := json.MarshalIndent(oldNodePools, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Old node pools: %s", string(pretty))
	pretty, err = json.MarshalIndent(newNodePools, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] New node pools: %s", string(pretty))

	// Determine if any node pools removed from the configuration.
	// Then downscale those pools and delete.
	for _, op := range oldNodePools {
		var found bool
		for _, np := range newNodePools {
			if op.Name == np.Name && op.Flavor == np.Flavor && op.Image == np.Image && (np.AvailabilityZone == "" || op.AvailabilityZone == np.AvailabilityZone) {
				tmp := np
				// copy previously "computed" AZ
				if np.AvailabilityZone == "" {
					tmp.AvailabilityZone = op.AvailabilityZone
				}
				poolsToKeep = append(poolsToKeep, tmp)
				found = true
			}
		}

		if !found {
			tmp := op
			tmp.Size = 0
			poolsToDelete = append(poolsToDelete, tmp)
		}
	}

	pretty, _ = json.MarshalIndent(poolsToKeep, "", "  ")
	log.Printf("[DEBUG] Keep node pools: %s", string(pretty))
	pretty, _ = json.MarshalIndent(poolsToDelete, "", "  ")
	log.Printf("[DEBUG] Downscale node pools: %s", string(pretty))

	if len(poolsToDelete) > 0 {
		// downscale
		cluster.Spec.NodePools = append(poolsToKeep, poolsToDelete...)
		err = kubernikusUpdateAndWait(ctx, klient, cluster, target, pending, timeout)
		if err != nil {
			return err
		}
	}

	// delete old
	cluster.Spec.NodePools = poolsToKeep
	err = kubernikusUpdateAndWait(ctx, klient, cluster, target, pending, timeout)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(poolsToKeep, newNodePools) {
		// create new
		cluster.Spec.NodePools = newNodePools
		err = kubernikusUpdateAndWait(ctx, klient, cluster, target, pending, timeout)
		if err != nil {
			return err
		}
	}

	return nil
}

func kubernikusHandleErrorV1(msg string, err error) error {
	switch res := err.(type) {
	case *operations.TerminateClusterDefault:
		return fmt.Errorf("%s: %s", msg, res.Payload.Message)
	case error:
		return fmt.Errorf("%s: %v", msg, err)
	}
	return err
}

func kubernikusUpdateAndWait(ctx context.Context, klient *kubernikus, cluster *models.Kluster, target string, pending []string, timeout time.Duration) error {
	_, err := klient.UpdateCluster(operations.NewUpdateClusterParams().WithName(cluster.Name).WithBody(cluster), klient.authFunc())
	if err != nil {
		return kubernikusHandleErrorV1("Error updating cluster", err)
	}

	err = kubernikusWaitForClusterV1(ctx, klient, cluster.Name, target, pending, timeout)
	if err != nil {
		return kubernikusHandleErrorV1("Error waiting for cluster node pools Running state", err)
	}

	return nil
}

func getCredentials(klient *kubernikus, name string, creds string) (string, []map[string]string, error) {
	var err error
	var kubeConfig []map[string]string
	var crt *x509.Certificate

	if creds == "" {
		creds, kubeConfig, err = downloadCredentials(klient, name)
		if err != nil {
			return "", nil, err
		}
	} else {
		kubeConfig, crt, err = flattenKubernetesClusterKubeConfig(creds)
		if err != nil {
			return "", nil, err
		}

		// Check so that the certificate is valid now
		now := time.Now()
		if now.Before(crt.NotBefore) || now.After(crt.NotAfter) {
			log.Printf("[DEBUG] The Kubernikus certificate is not valid")
			creds, kubeConfig, err = downloadCredentials(klient, name)
			if err != nil {
				return "", nil, err
			}
		}
	}

	return creds, kubeConfig, nil
}

func flattenKubernetesClusterKubeConfig(creds string) ([]map[string]string, *x509.Certificate, error) {
	var cfg clientcmdapi.Config
	values := make(map[string]string)
	var crt *x509.Certificate

	err := yaml.Unmarshal([]byte(creds), &cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal Kubernikus kubeconfig: %s", err)
	}

	for _, v := range cfg.Clusters {
		values["host"] = v.Cluster.Server
		values["cluster_ca_certificate"] = base64.StdEncoding.EncodeToString(v.Cluster.CertificateAuthorityData)
	}

	for _, v := range cfg.AuthInfos {
		values["username"] = v.Name
		values["client_certificate"] = base64.StdEncoding.EncodeToString(v.AuthInfo.ClientCertificateData)
		values["client_key"] = base64.StdEncoding.EncodeToString(v.AuthInfo.ClientKeyData)

		// parse certificate date
		pem, _ := pem.Decode(v.AuthInfo.ClientCertificateData)
		if pem == nil {
			return nil, nil, fmt.Errorf("failed to decode PEM")
		}
		crt, err = x509.ParseCertificate(pem.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse Kubernikus certificate %s", err)
		}
		values["not_before"] = crt.NotBefore.Format(time.RFC3339)
		values["not_after"] = crt.NotAfter.Format(time.RFC3339)
	}

	if crt == nil {
		return nil, nil, fmt.Errorf("failed to get Kubernikus kubeconfig credentials %s", err)
	}

	return []map[string]string{values}, crt, nil
}

func kubernikusFlattenOIDCV1(oidc *models.OIDC) []map[string]string {
	values := make(map[string]string)

	if oidc == nil || *oidc == (models.OIDC{}) {
		return nil
	}

	values["client_id"] = oidc.ClientID
	values["issuer_url"] = oidc.IssuerURL

	return []map[string]string{values}
}

func kubernikusExpandOIDCV1(raw any) *models.OIDC {
	v := raw.([]any)

	if len(v) == 0 {
		return nil
	}

	if len(v) > 1 {
		log.Printf("[WARN] More than one OIDC configuration found, using the first one.")
	}

	if v[0] == nil {
		return nil
	}

	if m, ok := v[0].(map[string]any); ok {
		res := &models.OIDC{
			ClientID:  m["client_id"].(string),
			IssuerURL: m["issuer_url"].(string),
		}

		if res.ClientID == "" || res.IssuerURL == "" {
			return nil
		}

		return res
	}

	return nil
}

func downloadCredentials(klient *kubernikus, name string) (string, []map[string]string, error) {
	credentials, err := klient.GetClusterCredentials(operations.NewGetClusterCredentialsParams().WithName(name), klient.authFunc())
	if err != nil {
		return "", nil, fmt.Errorf("failed to download Kubernikus kubeconfig: %s", err)
	}

	kubeConfig, _, err := flattenKubernetesClusterKubeConfig(credentials.Payload.Kubeconfig)
	if err != nil {
		return "", nil, err
	}

	return credentials.Payload.Kubeconfig, kubeConfig, nil
}

func verifySupportedKubernetesVersion(klient *kubernikus, version string) error {
	if info, err := klient.Info(nil); err != nil {
		return fmt.Errorf("failed to check supported Kubernetes versions: %s", err)
	} else if !strSliceContains(info.Payload.AvailableClusterVersions, version) {
		return fmt.Errorf("kubernikus doesn't support %q Kubernetes version, supported versions: %q", version, info.Payload.AvailableClusterVersions)
	}
	return nil
}
