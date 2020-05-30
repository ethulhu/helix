package upnp

import (
	"context"
	"net"
	"net/url"

	"github.com/ethulhu/helix/upnp/ssdp"
)

// DiscoverURLs discovers UPnP device manifest URLs using SSDP on the local network.
// It returns all valid URLs it finds, a slice of errors from invalid SSDP responses, and an error with the actual connection itself.
func DiscoverURLs(ctx context.Context, urn URN, iface *net.Interface) ([]*url.URL, []error, error) {
	return ssdp.DiscoverURLs(ctx, urn, iface)
}

// DiscoverURLs discovers UPnP device manifest URLs using SSDP on the local network.
// It returns all valid URLs it finds, a slice of errors from invalid SSDP responses, and an error with the actual connection itself.
func DiscoverDevices(ctx context.Context, urn URN, iface *net.Interface) ([]*Device, []error, error) {
	return ssdp.Discover(ctx, urn, iface)
}
