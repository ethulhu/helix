package contentdirectory

import (
	"context"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

type (
	Client interface {
		Name() string
		Browse(context.Context, upnpav.Object) ([]upnpav.Container, []upnpav.Item, error)
		SearchCapabilities(context.Context) ([]string, error)
	}
)

const (
	Version1 = ssdp.URN("urn:schemas-upnp-org:service:ContentDirectory:1")
	Version2 = ssdp.URN("urn:schemas-upnp-org:service:ContentDirectory:2")
	Version3 = ssdp.URN("urn:schemas-upnp-org:service:ContentDirectory:3")
)

const (
	Root = upnpav.Object("0")
)
