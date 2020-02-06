package contentdirectory

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

type (
	client struct {
		name       string
		soapClient soap.Client
	}
)

func Discover(ctx context.Context) ([]Client, []error, error) {
	// TODO: maybe discover all devices, then build a client around which version it has?
	devices, errs, err := ssdp.DiscoverDevices(ctx, Version1)

	var clients []Client
	for _, device := range devices {
		soapClient, ok := device.Client(Version1)
		if !ok {
			// TODO: expand this.
			errs = append(errs, fmt.Errorf("could not find ContentDirectory client"))
			continue
		}
		clients = append(clients, NewClient(device.Name(), soapClient))
	}
	return clients, errs, err
}

func NewClient(name string, soapClient soap.Client) Client {
	return &client{
		name:       name,
		soapClient: soapClient,
	}
}

func (c *client) Name() string {
	return c.name
}

func (c *client) call(ctx context.Context, method string, input, output interface{}) error {
	return c.soapClient.Call(ctx, string(Version1), method, input, output)
}

func (c *client) Browse(ctx context.Context, object upnpav.Object) ([]upnpav.Container, []upnpav.Item, error) {
	req := browseRequest{
		Object:     object,
		BrowseFlag: browseDirectChildren,
	}

	rsp := browseResponse{}
	if err := c.call(ctx, "Browse", req, &rsp); err != nil {
		return nil, nil, fmt.Errorf("could not perform Browse request: %w", err)
	}

	didl := upnpav.DIDL{}
	if err := xml.Unmarshal(rsp.Result, &didl); err != nil {
		return nil, nil, fmt.Errorf("could not unmarshal result: %w", err)
	}

	return didl.Containers, didl.Items, nil
}

func (c *client) SearchCapabilities(ctx context.Context) ([]string, error) {
	rsp := getSearchCapabilitiesResponse{}
	if err := c.call(ctx, "GetSearchCapabilities", nil, &rsp); err != nil {
		return nil, fmt.Errorf("could not get search capabilities: %w", err)
	}

	return strings.Split(rsp.Capabilities, ","), nil
}
