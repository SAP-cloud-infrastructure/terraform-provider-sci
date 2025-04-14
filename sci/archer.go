package sci

import (
	"fmt"
	"log"
	"net/url"
	"reflect"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/gophercloud/gophercloud/v2"
	osClient "github.com/gophercloud/utils/v2/client"
	"github.com/sapcc/archer/client"
)

type archer struct {
	client.Archer
	provider *gophercloud.ProviderClient
}

func newArcherV1(c *Config, eo gophercloud.EndpointOpts) (*archer, error) {
	var err error
	var endpoint string
	var aurl *url.URL

	if v, ok := c.EndpointOverrides["endpoint-services"]; ok {
		if e, ok := v.(string); ok && e != "" {
			endpoint = e
		}
	}

	if endpoint == "" && !reflect.DeepEqual(eo, gophercloud.EndpointOpts{}) {
		eo.ApplyDefaults("endpoint-services")
		endpoint, err = c.OsClient.EndpointLocator(eo)
		if err != nil {
			return nil, err
		}
	}

	if aurl, err = url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("parsing the Archer URL failed: %s", err)
	}

	transport := httptransport.New(aurl.Host, aurl.EscapedPath(), []string{aurl.Scheme})

	if v, ok := c.OsClient.HTTPClient.Transport.(*osClient.RoundTripper); ok && v.Logger != nil {
		// enable JSON debug for Archer
		transport.SetLogger(logger{"Archer"})
		transport.Debug = true
	}

	operations := client.New(transport, strfmt.Default)

	return &archer{*operations, c.OsClient}, nil
}

func (a *archer) authFunc() runtime.ClientAuthInfoWriterFunc {
	return runtime.ClientAuthInfoWriterFunc(
		func(req runtime.ClientRequest, reg strfmt.Registry) error {
			err := req.SetHeaderParam("X-Auth-Token", a.provider.Token())
			if err != nil {
				log.Printf("[DEBUG] Kubernikus auth func cannot set X-Auth-Token header value: %v", err)
			}
			err = req.SetHeaderParam("User-Agent", a.provider.UserAgent.Join())
			if err != nil {
				log.Printf("[DEBUG] Kubernikus auth func cannot set User-Agent header value: %v", err)
			}
			return nil
		})
}
