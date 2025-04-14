package sci

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/utils/v2/openstack/clientconfig"
	"github.com/sapcc/andromeda/client"
	"github.com/sapcc/gophercloud-sapcc/v2/clients"
)

func (c *Config) kubernikusV1Client(ctx context.Context, region string, isAdmin bool) (*kubernikus, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	serviceType := "kubernikus"
	if isAdmin {
		serviceType = "kubernikus-kubernikus"
	}

	return newKubernikusV1(c, gophercloud.EndpointOpts{
		Type:         serviceType,
		Region:       c.DetermineRegion(region),
		Availability: clientconfig.GetEndpointType(c.EndpointType),
	})
}

func (c *Config) andromedaV1Client(ctx context.Context, region string) (*client.Andromeda, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	return newAndromedaV1(c, gophercloud.EndpointOpts{
		Type:         "gtm",
		Region:       c.DetermineRegion(region),
		Availability: clientconfig.GetEndpointType(c.EndpointType),
	})
}

func (c *Config) archerV1Client(ctx context.Context, region string) (*archer, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	return newArcherV1(c, gophercloud.EndpointOpts{
		Type:         "endpoint-services",
		Region:       c.DetermineRegion(region),
		Availability: clientconfig.GetEndpointType(c.EndpointType),
	})
}

func (c *Config) arcV1Client(ctx context.Context, region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(ctx, clients.NewArcV1, region, "arc")
}

func (c *Config) automationV1Client(ctx context.Context, region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(ctx, clients.NewAutomationV1, region, "automation")
}

func (c *Config) billingClient(ctx context.Context, region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(ctx, clients.NewBilling, region, "sapcc-billing")
}
