// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package ssdp

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		manifest Document
		want     string
	}{
		{
			manifest: Document{
				SpecVersion: SpecVersion{
					Major: 1,
					Minor: 2,
				},
				Device: Device{
					DeviceType:   "foo",
					FriendlyName: "Foo (bar)",
				},
			},
			want: `<root xmlns="urn:schemas-upnp-org:device-1-0">
  <specVersion>
    <major>1</major>
    <minor>2</minor>
  </specVersion>
  <device>
    <deviceType>foo</deviceType>
    <friendlyName>Foo (bar)</friendlyName>
    <deviceList></deviceList>
    <iconList></iconList>
    <serviceList></serviceList>
  </device>
</root>`,
		},
	}

	for i, tt := range tests {
		bytes, err := xml.MarshalIndent(tt.manifest, "", "  ")
		if err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}
		got := string(bytes)

		if got != tt.want {
			t.Errorf("[%d]: got:\n\n%v\n\nwanted:\n\n%v", i, got, tt.want)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		raw  string
		want Document
	}{
		{
			raw: `<?xml version="1.0" encoding="utf-8"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
  <specVersion>
    <major>1</major>
    <minor>1</minor>
  </specVersion>
  <device>
    <deviceType>urn:schemas-upnp-org:device:MediaRenderer:1</deviceType>
    <friendlyName>UpMpd (valkyrie)</friendlyName>
    <manufacturer>Cats</manufacturer>
    <manufacturerURL>https://cats.cats</manufacturerURL>
    <modelDescription>UPnP thingy</modelDescription>
    <modelName>UpMPD</modelName>
    <modelNumber>42</modelNumber>
    <modelURL>https://framagit.org/medoc92/upmpdcli/code/</modelURL>
    <serialNumber>2020.05</serialNumber>
    <presentationURL>foo</presentationURL>    <UDN>udn:foo</UDN>
    <serviceList>
<service>
 <serviceType>urn:schemas-upnp-org:service:AVTransport:1</serviceType>
 <serviceId>urn:upnp-org:serviceId:AVTransport</serviceId>
 <SCPDURL>/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/urn-schemas-upnp-org-service-AVTransport-1.xml</SCPDURL>
 <controlURL>/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/ctl-urn-schemas-upnp-org-service-AVTransport-1</controlURL>
 <eventSubURL>/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/evt-urn-schemas-upnp-org-service-AVTransport-1</eventSubURL>
</service>
<service>
 <serviceType>urn:schemas-upnp-org:service:RenderingControl:1</serviceType>
 <serviceId>urn:upnp-org:serviceId:RenderingControl</serviceId>
 <SCPDURL>/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/urn-schemas-upnp-org-service-RenderingControl-1.xml</SCPDURL>
 <controlURL>/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/ctl-urn-schemas-upnp-org-service-RenderingControl-1</controlURL>
 <eventSubURL>/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/evt-urn-schemas-upnp-org-service-RenderingControl-1</eventSubURL>
</service>
    </serviceList>
  </device>
</root>`,
			want: Document{
				XMLName: xml.Name{
					Local: "root",
					Space: xmlns,
				},
				SpecVersion: SpecVersion{
					Major: 1,
					Minor: 1,
				},
				Device: Device{
					UDN:              "udn:foo",
					DeviceType:       "urn:schemas-upnp-org:device:MediaRenderer:1",
					FriendlyName:     "UpMpd (valkyrie)",
					Manufacturer:     "Cats",
					ManufacturerURL:  "https://cats.cats",
					ModelDescription: "UPnP thingy",
					ModelName:        "UpMPD",
					ModelNumber:      "42",
					ModelURL:         "https://framagit.org/medoc92/upmpdcli/code/",
					SerialNumber:     "2020.05",
					PresentationURL:  "foo",
					Services: []Service{
						{
							ServiceType: "urn:schemas-upnp-org:service:AVTransport:1",
							ServiceID:   "urn:upnp-org:serviceId:AVTransport",
							SCPDURL:     "/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/urn-schemas-upnp-org-service-AVTransport-1.xml",
							ControlURL:  "/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/ctl-urn-schemas-upnp-org-service-AVTransport-1",
							EventSubURL: "/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/evt-urn-schemas-upnp-org-service-AVTransport-1",
						},
						{
							ServiceType: "urn:schemas-upnp-org:service:RenderingControl:1",
							ServiceID:   "urn:upnp-org:serviceId:RenderingControl",
							SCPDURL:     "/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/urn-schemas-upnp-org-service-RenderingControl-1.xml",
							ControlURL:  "/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/ctl-urn-schemas-upnp-org-service-RenderingControl-1",
							EventSubURL: "/uuid-42f18105-7062-617c-5225-bc5ff4ed7a1e/evt-urn-schemas-upnp-org-service-RenderingControl-1",
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		got := Document{}
		if err := xml.Unmarshal([]byte(tt.raw), &got); err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}

		if !reflect.DeepEqual(got, tt.want) {
			gotBytes, _ := xml.MarshalIndent(got, "", "  ")
			wantBytes, _ := xml.MarshalIndent(tt.want, "", "  ")

			t.Errorf("[%d]: got:\n\n%s\n\nwanted:\n\n%s", i, gotBytes, wantBytes)
		}
	}
}