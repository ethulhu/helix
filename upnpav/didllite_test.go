// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"reflect"
	"testing"
)

func TestParseDIDLLite(t *testing.T) {
	tests := []struct {
		raw     string
		want    *DIDLLite
		wantErr error
	}{
		{
			raw: `<?xml version="1.0" encoding="UTF-8"?><DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/" xmlns:dlna="urn:schemas-dlna-org:metadata-1-0/"><item restricted="1" searchable="0"><res protocolInfo="http-get:*:audio/mpeg:*">http://192.168.16.4:8200/MediaItems/36.mp3</res></item></DIDL-Lite>`,
			want: &DIDLLite{
				Items: []Item{{
					Restricted: true,
					Searchable: false,
					Resources: []Resource{{
						URI: "http://192.168.16.4:8200/MediaItems/36.mp3",
						ProtocolInfo: &ProtocolInfo{
							Protocol:       ProtocolHTTP,
							Network:        "*",
							ContentFormat:  "audio/mpeg",
							AdditionalInfo: "*",
						},
					}},
				}},
			},
		},

		{
			raw: `
<DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/"
xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/"
xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/"
xmlns:dlna="urn:schemas-dlna-org:metadata-1-0/">
  <container id="64" parentID="0" restricted="1" searchable="1" childCount="4">
    <dc:title>Browse Folders</dc:title>
    <upnp:class>object.container.storageFolder</upnp:class>
    <upnp:storageUsed>-1</upnp:storageUsed>
  </container>
  <container id="1" parentID="0" restricted="0" searchable="0" childCount="7">
    <dc:title>Music</dc:title>
    <upnp:class>object.container.storageFolder</upnp:class>
    <upnp:storageUsed>-1</upnp:storageUsed>
  </container>
  <item id="72" parentID="4" restricted="0" searchable="0">
    <res protocolInfo="http-get:*:audio/mpeg:*" colorDepth="3">http://mew/purr.mp3</res>
    <res protocolInfo="http-get:*:video/mp4:*" resolution="480x360">http://mew/purr.mp4</res>
  </item>
</DIDL-Lite>
`,
			want: &DIDLLite{
				Containers: []Container{
					{
						ID:               ObjectID("64"),
						Parent:           ObjectID("0"),
						Restricted:       true,
						Searchable:       true,
						Title:            "Browse Folders",
						Class:            StorageFolder,
						ChildCount:       4,
						StorageUsedBytes: -1,
					},
					{
						ID:               ObjectID("1"),
						Parent:           ObjectID("0"),
						Restricted:       false,
						Searchable:       false,
						Title:            "Music",
						Class:            StorageFolder,
						ChildCount:       7,
						StorageUsedBytes: -1,
					},
				},
				Items: []Item{
					{
						ID:         ObjectID("72"),
						Parent:     ObjectID("4"),
						Restricted: false,
						Searchable: false,
						Resources: []Resource{
							{
								URI: "http://mew/purr.mp3",
								ProtocolInfo: &ProtocolInfo{
									Protocol:       ProtocolHTTP,
									Network:        "*",
									ContentFormat:  "audio/mpeg",
									AdditionalInfo: "*",
								},
								ColorDepth: 3,
							},
							{
								URI: "http://mew/purr.mp4",
								ProtocolInfo: &ProtocolInfo{
									Protocol:       ProtocolHTTP,
									Network:        "*",
									ContentFormat:  "video/mp4",
									AdditionalInfo: "*",
								},
								Resolution: &Resolution{
									Width:  480,
									Height: 360,
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		got, gotErr := ParseDIDLLite(tt.raw)
		if !reflect.DeepEqual(tt.wantErr, gotErr) {
			t.Errorf("[%d]: expected error %v, got %v", i, tt.wantErr, gotErr)
		}
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("[%d]: got:\n\n%v\n\nwant:\n\n%v\n\n", i, got, tt.want)
		}
	}
}

func TestMarshalDIDLLite(t *testing.T) {
	tests := []struct {
		didllite *DIDLLite
		want     string
	}{
		{
			didllite: &DIDLLite{},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<DIDL-Lite xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns:dlna="urn:schemas-dlna-org:metadata-1-0/"></DIDL-Lite>`,
		},
		{
			didllite: &DIDLLite{
				Containers: []Container{
					{
						ID:               ObjectID("64"),
						Parent:           ObjectID("0"),
						Restricted:       false,
						Searchable:       true,
						Title:            "Browse Folders",
						Class:            StorageFolder,
						ChildCount:       4,
						StorageUsedBytes: -1,
					},
					{
						ID:               ObjectID("1"),
						Parent:           ObjectID("0"),
						Restricted:       true,
						Searchable:       false,
						Title:            "Music",
						Class:            StorageFolder,
						ChildCount:       7,
						StorageUsedBytes: -1,
					},
				},
			},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<DIDL-Lite xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns:dlna="urn:schemas-dlna-org:metadata-1-0/">
  <container id="64" parentID="0" restricted="0" searchable="1" childCount="4">
    <dc:title>Browse Folders</dc:title>
    <upnp:class>object.container.storageFolder</upnp:class>
    <upnp:storageUsed>-1</upnp:storageUsed>
  </container>
  <container id="1" parentID="0" restricted="1" searchable="0" childCount="7">
    <dc:title>Music</dc:title>
    <upnp:class>object.container.storageFolder</upnp:class>
    <upnp:storageUsed>-1</upnp:storageUsed>
  </container>
</DIDL-Lite>`,
		},
		{
			didllite: &DIDLLite{
				Items: []Item{
					{
						ID:         ObjectID("69"),
						Parent:     ObjectID("12"),
						Restricted: false,
						Searchable: true,
						Title:      "hello",
						Resources: []Resource{
							{
								URI: "http://mew/purr.mp3",
								ProtocolInfo: &ProtocolInfo{
									Protocol:      ProtocolHTTP,
									ContentFormat: "audio/mpeg",
								},
								BitsPerSecond: 128 * 1024,
							},
							{
								URI: "http://mew/purr.mp4",
								ProtocolInfo: &ProtocolInfo{
									Protocol:      ProtocolHTTP,
									ContentFormat: "video/mp4",
								},
								Resolution: &Resolution{
									Width:  480,
									Height: 360,
								},
							},
						},
					},
				},
			},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<DIDL-Lite xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns:dlna="urn:schemas-dlna-org:metadata-1-0/">
  <item id="69" parentID="12" restricted="0" searchable="1">
    <dc:title>hello</dc:title>
    <res protocolInfo="http-get:*:audio/mpeg:*" bitrate="131072">http://mew/purr.mp3</res>
    <res protocolInfo="http-get:*:video/mp4:*" resolution="480x360">http://mew/purr.mp4</res>
  </item>
</DIDL-Lite>`,
		},
	}

	for i, tt := range tests {
		got := tt.didllite.String()
		if tt.want != got {
			t.Errorf("[%d]: got:\n\n%+v\n\nwant:\n\n%+v", i, got, tt.want)
		}
	}
}
