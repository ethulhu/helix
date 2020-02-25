package connectionmanager

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnpav"
)

type (
	client struct{ soap.Client }
)

func NewClient(soapClient soap.Client) Client {
	return &client{soapClient}
}

func (c *client) ProtocolInfo(ctx context.Context) ([]*upnpav.ProtocolInfo, []*upnpav.ProtocolInfo, error) {
	req := getProtocolInfoRequest{}
	rsp := getProtocolInfoResponse{}
	if err := c.Call(ctx, string(Version1), "GetProtocolInfo", req, &rsp); err != nil {
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
