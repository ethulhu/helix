package upnpav

import (
	"reflect"
	"testing"
)

func TestParseDIDLLite(t *testing.T) {
	tests := []struct {
		raw     string
		want    *DIDL
		wantErr error
	}{
		{
			raw: `<?xml version="1.0" encoding="utf-8"?><DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/" xmlns:dlna="urn:schemas-dlna-org:metadata-1-0/"><item restricted="1" searchable="0"><res protocolInfo="http-get:*:audio/mpeg:*">http://192.168.16.4:8200/MediaItems/36.mp3</res></item></DIDL-Lite>`,
			want: &DIDL{
				Items: []Item{{
					Restricted: "1",
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
  <container id="64" parentID="0" restricted="1" searchable="1"
  childCount="4">
    <dc:title>Browse Folders</dc:title>
    <upnp:class>object.container.storageFolder</upnp:class>
    <upnp:storageUsed>-1</upnp:storageUsed>
  </container>
  <container id="1" parentID="0" restricted="1" searchable="1"
  childCount="7">
    <dc:title>Music</dc:title>
    <upnp:class>object.container.storageFolder</upnp:class>
    <upnp:storageUsed>-1</upnp:storageUsed>
  </container>
</DIDL-Lite>
`,
			want: &DIDL{
				Containers: []Container{
					{
						ID:          Object("64"),
						ParentID:    Object("0"),
						Restricted:  "1",
						Searchable:  "1",
						ChildCount:  "4",
						Title:       "Browse Folders",
						Class:       "object.container.storageFolder",
						StorageUsed: "-1",
					},
					{
						ID:          Object("1"),
						ParentID:    Object("0"),
						Restricted:  "1",
						Searchable:  "1",
						ChildCount:  "7",
						Title:       "Music",
						Class:       "object.container.storageFolder",
						StorageUsed: "-1",
					},
				},
			},
		},
	}

	for i, tt := range tests {
		got, gotErr := ParseDIDL([]byte(tt.raw))
		if !reflect.DeepEqual(tt.wantErr, gotErr) {
			t.Errorf("[%d]: expected error %v, got %v", i, tt.wantErr, gotErr)
		}
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("[%d]: expected result:\n\n%+v\n\ngot:\n\n%+v", i, tt.want, got)
		}
	}
}
