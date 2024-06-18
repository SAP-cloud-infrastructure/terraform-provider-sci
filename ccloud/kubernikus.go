package ccloud

import (
	"fmt"
	"log"
	"net/url"
	"reflect"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/sapcc/kubernikus/pkg/api/client/operations"

	"github.com/gophercloud/gophercloud"
	osClient "github.com/gophercloud/utils/client"
)

type kubernikus struct {
	operations.ClientService
	provider *gophercloud.ProviderClient
}

func newKubernikusV1(c *Config, eo gophercloud.EndpointOpts) (*kubernikus, error) {
	var err error
	var endpoint string
	var kurl *url.URL

	if v, ok := c.EndpointOverrides["kubernikus"]; ok {
		if e, ok := v.(string); ok && e != "" {
			endpoint = e
		}
	}

	if endpoint == "" && !reflect.DeepEqual(eo, gophercloud.EndpointOpts{}) {
		eo.ApplyDefaults("kubernikus")
		endpoint, err = c.OsClient.EndpointLocator(eo)
		if err != nil {
			return nil, err
		}
	}

	if kurl, err = url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("Parsing the Kubernikus URL failed: %s", err)
	}

	transport := httptransport.New(kurl.Host, kurl.EscapedPath(), []string{kurl.Scheme})

	if v, ok := c.OsClient.HTTPClient.Transport.(*osClient.RoundTripper); ok && v.Logger != nil {
		// enable JSON debug for Kubernikus
		transport.SetLogger(logger{"Kubernikus"})
		transport.Debug = true
	}

	operations := operations.New(transport, strfmt.Default)

	return &kubernikus{operations, c.OsClient}, nil
}

func (k *kubernikus) authFunc() runtime.ClientAuthInfoWriterFunc {
	return runtime.ClientAuthInfoWriterFunc(
		func(req runtime.ClientRequest, reg strfmt.Registry) error {
			err := req.SetHeaderParam("X-Auth-Token", k.provider.Token())
			if err != nil {
				log.Printf("[DEBUG] Kubernikus auth func cannot set X-Auth-Token header value: %v", err)
			}
			err = req.SetHeaderParam("User-Agent", k.provider.UserAgent.Join())
			if err != nil {
				log.Printf("[DEBUG] Kubernikus auth func cannot set User-Agent header value: %v", err)
			}
			return nil
		})
}
