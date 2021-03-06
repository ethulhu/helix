// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/ethulhu/helix/logger"
	"github.com/ethulhu/helix/upnp/httpu"
	"github.com/ethulhu/helix/upnp/ssdp"
)

const (
	discoverMethod = "M-SEARCH"
	notifyMethod   = "NOTIFY"

	ssdpCacheControl = "max-age=300"
)

var (
	ssdpBroadcastAddr = &net.UDPAddr{
		IP:   net.IPv4(239, 255, 255, 250),
		Port: 1900,
	}

	discoverURL = &url.URL{Opaque: "*"}
)

// DiscoverURLs discovers UPnP device manifest URLs using SSDP on the local network.
// It returns all valid URLs it finds, a slice of errors from invalid SSDP responses, and an error with the actual connection itself.
func DiscoverURLs(ctx context.Context, urn URN, iface *net.Interface) ([]*url.URL, []error, error) {
	req := discoverRequest(ctx, urn)

	rsps, errs, err := httpu.Do(req, 3, iface)

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

// DiscoverDevices discovers UPnP devices using SSDP on the local network.
// It returns all valid URLs it finds, a slice of errors from invalid SSDP responses or UPnP device manifests, and an error with the actual connection itself.
func DiscoverDevices(ctx context.Context, urn URN, iface *net.Interface) ([]*Device, []error, error) {
	urls, errs, err := DiscoverURLs(ctx, urn, iface)

	var devices []*Device
	for _, manifestURL := range urls {
		rsp, err := http.Get(manifestURL.String())
		if err != nil {
			errs = append(errs, fmt.Errorf("could not GET manifest %v: %w", manifestURL, err))
			continue
		}
		bytes, _ := ioutil.ReadAll(rsp.Body)

		manifest := ssdp.Document{}
		if err := xml.Unmarshal(bytes, &manifest); err != nil {
			errs = append(errs, err)
			continue
		}

		device, err := newDevice(manifestURL, manifest)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		devices = append(devices, device)
	}
	return devices, errs, err
}

// BroadcastDevice broadcasts the presence of a UPnP Device, with its SSDP/SCPD served via HTTP at addr.
func BroadcastDevice(d *Device, url string, iface *net.Interface) error {
	conn, err := net.ListenMulticastUDP("udp", iface, ssdpBroadcastAddr)
	if err != nil {
		return fmt.Errorf("could not listen on %v: %v", ssdpBroadcastAddr, err)
	}
	defer conn.Close()

	log, _ := logger.FromContext(context.TODO())
	log.WithField("httpu.listener", ssdpBroadcastAddr).Info("serving HTTPU")
	s := &httpu.Server{
		Handler: func(r *http.Request) []httpu.Response {
			switch r.Method {
			case discoverMethod:
				return handleDiscover(r, d, url)
			case notifyMethod:
				// TODO: handleNotify()
				return nil
			default:
				log, _ := logger.FromContext(r.Context())
				log.Warning("unknown method")
				return nil
			}
		},
	}
	return s.Serve(conn)
}

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

func handleDiscover(r *http.Request, d *Device, url string) []httpu.Response {
	log, _ := logger.FromContext(r.Context())

	if r.Header.Get("Man") != `"ssdp:discover"` {
		log.Warning("request lacked correct MAN header")
		return nil
	}

	st := URN(r.Header.Get("St"))

	ok := false
	for _, urn := range d.allURNs() {
		ok = ok || urn == st
	}
	if st == All || ok {
		responses := []httpu.Response{{
			"CACHE-CONTROL": ssdpCacheControl,
			"EXT":           "",
			"LOCATION":      url,
			"SERVER":        fmt.Sprintf("%s %s", d.ModelName, d.ModelNumber),
			"ST":            d.UDN,
			"USN":           d.UDN,
		}}
		for _, urn := range d.allURNs() {
			responses = append(responses, httpu.Response{
				"CACHE-CONTROL": ssdpCacheControl,
				"EXT":           "",
				"LOCATION":      url,
				"SERVER":        fmt.Sprintf("%s %s", d.ModelName, d.ModelNumber),
				"ST":            string(urn),
				"USN":           fmt.Sprintf("%s::%s", d.UDN, urn),
			})
		}
		return responses
	}
	return nil
}
