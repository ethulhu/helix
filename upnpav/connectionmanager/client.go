package connectionmanager

import (
	"context"
	"fmt"
	"log"
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

func NewClient(name string, soapClient soap.Client) Client {
	return &client{
		name:       name,
		soapClient: soapClient,
	}
}

func Discover(ctx context.Context) ([]Client, []error, error) {
	devices, errs, err := ssdp.DiscoverDevices(ctx, Version1)

	var clients []Client
	for _, device := range devices {
		soapClient, ok := device.Client(Version1)
		if !ok {
			// TODO: expand this.
			errs = append(errs, fmt.Errorf("could not find ConnectionManager client"))
			continue
		}
		clients = append(clients, NewClient(device.Name(), soapClient))
	}
	return clients, errs, err
}

func (c *client) Name() string {
	return c.name
}

func (c *client) ProtocolInfo(ctx context.Context) ([]*upnpav.ProtocolInfo, []*upnpav.ProtocolInfo, error) {
	req := getProtocolInfoRequest{}
	rsp := getProtocolInfoResponse{}
	if err := c.soapClient.Call(ctx, string(Version1), "GetProtocolInfo", req, &rsp); err != nil {
		return nil, nil, fmt.Errorf("could not get protocol info: %w", err)
	}

	var sources []*upnpav.ProtocolInfo
	for _, source := range strings.Split(rsp.Source, ",") {
		if source == "" {
			continue
		}
		protocolInfo, err := upnpav.ParseProtocolInfo(source)
		if err != nil {
			// TODO: do something proper here.
			log.Printf("o no: %v", err)
		}
		sources = append(sources, protocolInfo)
	}

	var sinks []*upnpav.ProtocolInfo
	for _, sink := range strings.Split(rsp.Sink, ",") {
		if sink == "" {
			continue
		}
		protocolInfo, err := upnpav.ParseProtocolInfo(sink)
		if err != nil {
			// TODO: do something proper here.
			log.Printf("o no: %v", err)
		}
		sinks = append(sinks, protocolInfo)
	}
	return sources, sinks, nil
}
