// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/ethulhu/helix/upnp/httpu"
)

func TestHandleDiscover(t *testing.T) {
	tests := []struct {
		req    *http.Request
		device *Device
		url    string
		want   []httpu.Response
	}{
		{
			req: &http.Request{
				Method: "M-SEARCH",
				Host:   "239.255.255.250:1900",
				URL:    &url.URL{Opaque: "*"},
				Header: http.Header{
					"Man": {`"ssdp:discover"`},
					"Mx":  {"2"},
					"St":  {"ssdp:all"},
				},
			},
			device: &Device{
				DeviceType: DeviceType("device-type"),
				UDN:        "device-id",
			},
			url: "http://1.2.3.4:8000/",
			want: []httpu.Response{
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "device-id",
					"USN":           "device-id",
				},
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "device-type",
					"USN":           "device-id::device-type",
				},
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "upnp:rootdevice",
					"USN":           "device-id::upnp:rootdevice",
				},
			},
		},
		{
			req: &http.Request{
				Method: "M-SEARCH",
				Host:   "239.255.255.250:1900",
				URL:    &url.URL{Opaque: "*"},
				Header: http.Header{
					"Man": {`"ssdp:discover"`},
					"Mx":  {"2"},
					"St":  {"device-type"},
				},
			},
			device: &Device{
				DeviceType: DeviceType("device-type"),
				UDN:        "device-id",
			},
			url: "http://1.2.3.4:8000/",
			want: []httpu.Response{
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "device-id",
					"USN":           "device-id",
				},
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "device-type",
					"USN":           "device-id::device-type",
				},
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "upnp:rootdevice",
					"USN":           "device-id::upnp:rootdevice",
				},
			},
		},
		{
			req: &http.Request{
				Method: "M-SEARCH",
				Host:   "239.255.255.250:1900",
				URL:    &url.URL{Opaque: "*"},
				Header: http.Header{
					"Man": {`"ssdp:discover"`},
					"Mx":  {"2"},
					"St":  {"service-urn"},
				},
			},
			device: &Device{
				DeviceType: DeviceType("device-type"),
				UDN:        "device-id",
				serviceByURN: map[URN]service{
					"service-urn": service{},
				},
			},
			url: "http://1.2.3.4:8000/",
			want: []httpu.Response{
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "device-id",
					"USN":           "device-id",
				},
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "service-urn",
					"USN":           "device-id::service-urn",
				},
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "device-type",
					"USN":           "device-id::device-type",
				},
				{
					"CACHE-CONTROL": ssdpCacheControl,
					"EXT":           "",
					"LOCATION":      "http://1.2.3.4:8000/",
					"SERVER":        " ",
					"ST":            "upnp:rootdevice",
					"USN":           "device-id::upnp:rootdevice",
				},
			},
		},
		{
			req: &http.Request{
				Method: "M-SEARCH",
				Host:   "239.255.255.250:1900",
				URL:    &url.URL{Opaque: "*"},
				Header: http.Header{
					"Man": {`"ssdp:discover"`},
					"Mx":  {"2"},
					"St":  {"twaddle"},
				},
			},
			device: &Device{
				DeviceType: DeviceType("meowpurr"),
				UDN:        "foobar",
				serviceByURN: map[URN]service{
					"tweedle": service{},
				},
			},
			url:  "http://1.2.3.4:8000/",
			want: nil,
		},
	}

	for i, tt := range tests {
		got := handleDiscover(tt.req, tt.device, tt.url)

		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]: got:\n\n%v\n\nwant:\n\n%v", i, got, tt.want)
		}
	}
}
