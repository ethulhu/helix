// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

import (
	"reflect"
	"testing"

	"github.com/ethulhu/helix/upnp/ssdp"
)

func TestIconSSDPIcon(t *testing.T) {
	tests := []struct {
		icon Icon
		want ssdp.Icon
	}{
		{
			icon: Icon{
				MIMEType: "video/mp4",
				Width:    10,
				Height:   12,
				Depth:    9,
				URL:      "/foo.wmv",
			},
			want: ssdp.Icon{
				MIMEType: "video/mp4",
				Width:    10,
				Height:   12,
				Depth:    9,
				URL:      "/foo.wmv",
			},
		},
		{
			icon: Icon{
				Width:  10,
				Height: 12,
				Depth:  9,
				URL:    "/foo.png",
			},
			want: ssdp.Icon{
				MIMEType: "image/png",
				Width:    10,
				Height:   12,
				Depth:    9,
				URL:      "/foo.png",
			},
		},
	}

	for i, tt := range tests {
		got := tt.icon.ssdpIcon()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]: got %v, want %v", i, got, tt.want)
		}
	}
}