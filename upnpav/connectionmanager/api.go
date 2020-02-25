package connectionmanager

import (
	"context"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

type (
	Client interface {
		ProtocolInfo(context.Context) ([]*upnpav.ProtocolInfo, []*upnpav.ProtocolInfo, error)
	}
)

const (
	Version1 = ssdp.URN("urn:schemas-upnp-org:service:ConnectionManager:1")
	Version2 = ssdp.URN("urn:schemas-upnp-org:service:ConnectionManager:2")
)
