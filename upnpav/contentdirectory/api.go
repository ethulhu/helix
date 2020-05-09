package contentdirectory

import (
	"context"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

type (
	Client interface {
		// BrowseMetadata shows information about a given object.
		BrowseMetadata(context.Context, upnpav.Object) (*upnpav.DIDL, error)

		// BrowseChildren lists the child objects of a given object.
		BrowseChildren(context.Context, upnpav.Object) (*upnpav.DIDL, error)

		// SearchCapabilities returns the search capabilities of the ContentDirectory service.
		SearchCapabilities(context.Context) ([]string, error)

		// Search queries the ContentDirectory service for objects under a given object that match a given criteria.
		Search(context.Context, upnpav.Object, search.Criteria) (*upnpav.DIDL, error)
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
