package upnpav

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseProtocolInfo(t *testing.T) {
	tests := []struct {
		raw     string
		want    *ProtocolInfo
		wantErr error
	}{
		{
			raw:     "",
			want:    nil,
			wantErr: fmt.Errorf("ProtocolInfo must have 4 parts, found 1"),
		},
		{
			raw:     "a:b:c:d:e",
			want:    nil,
			wantErr: fmt.Errorf("ProtocolInfo must have 4 parts, found 5"),
		},
		{
			raw: "http-get:*:audio/mpeg:*",
			want: &ProtocolInfo{
				Protocol:       ProtocolHTTP,
				Network:        "*",
				ContentFormat:  "audio/mpeg",
				AdditionalInfo: "*",
			},
			wantErr: nil,
		},
		{
			raw: "a:b:c:d",
			want: &ProtocolInfo{
				Protocol:       Protocol("a"),
				Network:        "b",
				ContentFormat:  "c",
				AdditionalInfo: "d",
			},
			wantErr: nil,
		},
	}

	for i, tt := range tests {
		got, gotErr := ParseProtocolInfo(tt.raw)
		if !reflect.DeepEqual(tt.wantErr, gotErr) {
			t.Errorf("[%d]: expected error %v, got %v", i, tt.wantErr, gotErr)
		}
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("[%d]: expected result %v, got %v", i, tt.want, got)
		}
	}
}

func TestProtocolInfoString(t *testing.T) {
	tests := []struct {
		protocolInfo ProtocolInfo
		want         string
	}{
		{
			protocolInfo: ProtocolInfo{
				Protocol:       ProtocolHTTP,
				Network:        "*",
				ContentFormat:  "audio/mpeg",
				AdditionalInfo: "*",
			},
			want: "http-get:*:audio/mpeg:*",
		},
		{
			protocolInfo: ProtocolInfo{
				Protocol:       Protocol("a"),
				Network:        "b",
				ContentFormat:  "c",
				AdditionalInfo: "d",
			},
			want: "a:b:c:d",
		},
	}

	for i, tt := range tests {
		got := tt.protocolInfo.String()
		if got != tt.want {
			t.Errorf("[%d]: expected result %v, got %v", i, tt.want, got)
		}
	}
}

func TestProtocolInfoForURI(t *testing.T) {
	tests := []struct {
		uri     string
		want    *ProtocolInfo
		wantErr error
	}{
		{
			uri: "http://mew.purr/meow.mp3",
			want: &ProtocolInfo{
				Protocol:       ProtocolHTTP,
				Network:        "*",
				ContentFormat:  "audio/mpeg",
				AdditionalInfo: "*",
			},
			wantErr: nil,
		},
		{
			uri: "http://mew.purr/meow.mkv",
			want: &ProtocolInfo{
				Protocol:       ProtocolHTTP,
				Network:        "*",
				ContentFormat:  "video/x-matroska",
				AdditionalInfo: "*",
			},
			wantErr: nil,
		},
		{
			uri:     "http://mew.purr/meow",
			want:    nil,
			wantErr: ErrUnknownMIMEType,
		},
	}

	for i, tt := range tests {
		got, gotErr := ProtocolInfoForURI(tt.uri)
		if !reflect.DeepEqual(tt.wantErr, gotErr) {
			t.Errorf("[%d]: expected error %v, got %v", i, tt.wantErr, gotErr)
		}
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("[%d]: expected result %v, got %v", i, tt.want, got)
		}
	}
}
