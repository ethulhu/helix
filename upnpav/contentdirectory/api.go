package contentdirectory

import (
	"context"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

type (
	Client interface {
		// BrowseMetadata shows information about a given object.
		BrowseMetadata(context.Context, upnpav.Object) (*upnpav.DIDL, error)
		// BrowseChildren lists the child objects of a given object.
		BrowseChildren(context.Context, upnpav.Object) (*upnpav.DIDL, error)
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
