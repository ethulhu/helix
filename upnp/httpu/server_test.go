package httpu

import (
	"reflect"
	"strings"
	"testing"
)

func TestSerializeResponse(t *testing.T) {
	tests := []struct {
		rsp  Response
		want string
	}{
		{
			rsp: Response{
				"LOCATION": "http://foo.com",
				"USN":      "uuid:6006::upnp:rootdevice",
				"EXT":      "",
			},
			want: `HTTP/1.1 200 OK
EXT:
LOCATION: http://foo.com
USN: uuid:6006::upnp:rootdevice
`,
		},
	}

	for i, tt := range tests {
		var want []byte
		for _, line := range strings.Split(tt.want, "\n") {
			want = append(want, []byte(line)...)
			want = append(want, []byte("\r\n")...)
		}

		got := tt.rsp.Bytes()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("[%d]: want:\n\n%s\n\ngot:\n\n%s", i, want, got)
		}
	}
}
