package ssdp

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/ethulhu/helix/upnp/httpu"
)

const (
	discoverMethod = "M-SEARCH"
)

var (
	ssdpBroadcastAddr = &net.UDPAddr{
		IP:   net.IPv4(239, 255, 255, 250),
		Port: 1900,
	}
	discoverURL = &url.URL{Opaque: "*"}
)

func discoverRequest(ctx context.Context, urn URN) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, discoverMethod, discoverURL.String(), http.NoBody)
	req.Host = ssdpBroadcastAddr.String()
	req.Header = http.Header{
		"MAN": {`"ssdp:discover"`},
		"MX":  {"2"},
		"ST":  {string(urn)},
	}
	return req
}

// DiscoverURLs discovers UPnP device manifest URLs using SSDP on the local network.
// It returns all valid URLs it finds, a slice of errors from invalid SSDP responses, and an error with the actual connection itself.
func DiscoverURLs(ctx context.Context, urn URN) ([]*url.URL, []error, error) {
	req := discoverRequest(ctx, urn)

	rsps, errs, err := httpu.Do(req, 3)

	locations := map[string]*url.URL{}
	for _, rsp := range rsps {
		location, err := rsp.Location()
		if err != nil {
			errs = append(errs, fmt.Errorf("could not find SSDP response Location: %w", err))
			continue
		}
		locations[location.String()] = location
	}

	var urls []*url.URL
	for _, location := range locations {
		urls = append(urls, location)
	}
	return urls, errs, err
}

// DiscoverURLs discovers UPnP devices using SSDP on the local network.
// It returns all valid URLs it finds, a slice of errors from invalid SSDP responses or UPnP device manifests, and an error with the actual connection itself.
func DiscoverDevices(ctx context.Context, urn URN) ([]Device, []error, error) {
	urls, errs, err := DiscoverURLs(ctx, urn)

	var devices []Device
	for _, manifestURL := range urls {
		rsp, err := http.Get(manifestURL.String())
		if err != nil {
			errs = append(errs, fmt.Errorf("could not GET manifest %v: %w", manifestURL, err))
			continue
		}
		bytes, _ := ioutil.ReadAll(rsp.Body)
		device, err := newDevice(manifestURL, bytes)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		devices = append(devices, device)
	}
	return devices, errs, err
}
