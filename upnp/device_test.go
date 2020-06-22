// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

import (
	"reflect"
	"testing"

	"github.com/ethulhu/helix/upnp/scpd"
	"github.com/ethulhu/helix/upnp/ssdp"
)

func TestDeviceManifest(t *testing.T) {
	tests := []struct {
		urns []URN
		ids  []ServiceID
		want ssdp.Document
	}{
		{
			urns: []URN{"hello"},
			ids:  []ServiceID{"goodbye"},
			want: ssdp.Document{
				SpecVersion: ssdp.Version,
				Device: ssdp.Device{
					FriendlyName:     "name",
					UDN:              "udn",
					Services: []ssdp.Service{
						{
							ServiceType: "hello",
							ServiceID:   "goodbye",
							SCPDURL:     "/hello",
							ControlURL:  "/hello",
							EventSubURL: "/hello",
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		d := &Device{
			Name: "name",
			UDN: "udn",
		}
		for i, urn := range tt.urns {
			d.Handle(urn, tt.ids[i], scpd.Document{}, nil)
		}

		got := d.manifest("/")
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]: got:\n\n%v\n\nwant:\n\n%v", i, got, tt.want)
		}
	}
}
