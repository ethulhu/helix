// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package soap

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestSerializeSOAPEnvelope(t *testing.T) {
	tests := []struct {
		input interface{}
		want  string
	}{
		{
			input: nil,
			want: `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body></s:Body></s:Envelope>`,
		},
		{
			input: struct {
				XMLName  xml.Name `xml:"https://mew.purr/cats GetNoises"`
				Adorable bool     `xml:"Adorable"`
			}{
				Adorable: true,
			},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body><GetNoises xmlns="https://mew.purr/cats"><Adorable>true</Adorable></GetNoises></s:Body></s:Envelope>`,
		},
	}

	for i, tt := range tests {
		got := serializeSOAPEnvelope(tt.input)
		want := []byte(tt.want)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("[%d]: got:\n\n%s\n\nwant:\n\n%s", i, got, want)
		}
	}
}

func TestDeserializeSOAPEnvelope(t *testing.T) {
	tests := []struct {
		raw     string
		want    interface{}
		wantErr error
	}{
		{
			raw: `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<s:Fault>
<faultcode>s:Client</faultcode>
<faultstring>UPnPError</faultstring>
<detail>
<UPnPError xmlns="urn:schemas-upnp-org:control-1-0">
<errorCode>501</errorCode>
<errorDescription>Action Failed</errorDescription>
</UPnPError>
</detail>
</s:Fault>
</s:Body>
</s:Envelope>`,
			want: nil,
			wantErr: &RemoteError{
				FaultCode:   FaultClient,
				FaultString: "UPnPError",
				Detail: `<UPnPError xmlns="urn:schemas-upnp-org:control-1-0">
<errorCode>501</errorCode>
<errorDescription>Action Failed</errorDescription>
</UPnPError>`,
			},
		},
	}

	for i, tt := range tests {
		var got interface{} = nil
		gotErr := deserializeSOAPEnvelope([]byte(tt.raw), got)

		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("[%d], got %v, want %v", i, got, tt.want)
		}
		if !reflect.DeepEqual(tt.wantErr, gotErr) {
			t.Errorf("[%d], got error %q, want error %q", i, gotErr, tt.wantErr)
		}
	}
}
