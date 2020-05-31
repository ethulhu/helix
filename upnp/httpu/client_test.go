// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package httpu

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestSerializeRequest(t *testing.T) {
	tests := []struct {
		req  *http.Request
		want string
	}{
		{
			req: &http.Request{
				Method: "M-SEARCH",
				Host:   "239.255.255.250:1900",
				URL:    &url.URL{Opaque: "*"},
				Header: http.Header{
					"MAN": {`"ssdp:discover"`},
					"MX":  {"2"},
					"ST":  {"ssdp:all"},
				},
			},
			want: `M-SEARCH * HTTP/1.1
HOST: 239.255.255.250:1900
MAN: "ssdp:discover"
MX: 2
ST: ssdp:all
`,
		},
	}

	for i, tt := range tests {
		var want []byte
		for _, line := range strings.Split(tt.want, "\n") {
			want = append(want, []byte(line)...)
			want = append(want, []byte("\r\n")...)
		}

		got := serializeRequest(tt.req)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("[%d]: want:\n\n%s\n\ngot:\n\n%s", i, want, got)
		}

	}

}
