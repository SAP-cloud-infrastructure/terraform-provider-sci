package sci

import (
	"fmt"
	"log"
	"net/url"
	"reflect"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/sapcc/andromeda/client"

	"github.com/gophercloud/gophercloud/v2"
	osClient "github.com/gophercloud/utils/v2/client"
)

func newAndromedaV1(c *Config, eo gophercloud.EndpointOpts) (*client.Andromeda, error) {
	var err error
	var endpoint string
	var aurl *url.URL

	if v, ok := c.EndpointOverrides["gtm"]; ok {
		if e, ok := v.(string); ok && e != "" {
			endpoint = e
		}
	}

	if endpoint == "" && !reflect.DeepEqual(eo, gophercloud.EndpointOpts{}) {
		eo.ApplyDefaults("gtm")
		endpoint, err = c.OsClient.EndpointLocator(eo)
		if err != nil {
			return nil, err
		}
	}

	if aurl, err = url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("parsing the Andromeda URL failed: %s", err)
	}

	transport := httptransport.New(aurl.Host, aurl.EscapedPath(), []string{aurl.Scheme})

	if v, ok := c.OsClient.HTTPClient.Transport.(*osClient.RoundTripper); ok && v.Logger != nil {
		// enable JSON debug for Andromeda
		transport.SetLogger(logger{"Andromeda"})
		transport.Debug = true
	}

	transport.DefaultAuthentication = runtime.ClientAuthInfoWriterFunc(
		func(req runtime.ClientRequest, reg strfmt.Registry) error {
			err := req.SetHeaderParam("X-Auth-Token", c.OsClient.Token())
			if err != nil {
				log.Printf("[DEBUG] Andromeda auth func cannot set X-Auth-Token header value: %v", err)
			}
			err = req.SetHeaderParam("User-Agent", c.OsClient.UserAgent.Join())
			if err != nil {
				log.Printf("[DEBUG] Andromeda auth func cannot set User-Agent header value: %v", err)
			}
			return nil
		})

	operations := client.New(transport, strfmt.Default)

	return operations, nil
}
