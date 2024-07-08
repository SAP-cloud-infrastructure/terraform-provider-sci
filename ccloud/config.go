package ccloud

import (
	"github.com/sapcc/andromeda/client"
	"github.com/sapcc/gophercloud-sapcc/clients"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

func (c *Config) kubernikusV1Client(region string, isAdmin bool) (*kubernikus, error) {
	if err := c.Authenticate(); err != nil {
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

func (c *Config) andromedaV1Client(region string) (*client.Andromeda, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	return newAndromedaV1(c, gophercloud.EndpointOpts{
		Type:         "gtm",
		Region:       c.DetermineRegion(region),
		Availability: clientconfig.GetEndpointType(c.EndpointType),
	})
}

func (c *Config) archerV1Client(region string) (*archer, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	return newArcherV1(c, gophercloud.EndpointOpts{
		Type:         "endpoint-services",
		Region:       c.DetermineRegion(region),
		Availability: clientconfig.GetEndpointType(c.EndpointType),
	})
}

func (c *Config) arcV1Client(region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(clients.NewArcV1, region, "arc")
}

func (c *Config) automationV1Client(region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(clients.NewAutomationV1, region, "automation")
}

func (c *Config) billingClient(region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(clients.NewBilling, region, "sapcc-billing")
}
