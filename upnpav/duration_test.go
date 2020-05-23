// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{
			raw:  "2:00:03",
			want: "2h3s",
		},
		{
			raw:  "0:40:03",
			want: "40m3s",
		},
	}

	for i, tt := range tests {
		want, err := time.ParseDuration(tt.want)
		if err != nil {
			t.Fatalf("[%d]: could not parse want %v: %v", i, tt.want, err)
		}

		got, err := ParseDuration(tt.raw)
		if err != nil {
			t.Fatalf("[%d]: could not parse raw %v: %v", i, tt.raw, err)
		}
		if got != want {
			t.Errorf("[%d]: got %v, wanted %v", i, got, want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration string
		want     string
	}{
		{
			duration: "2h3s",
			want:     "2:00:03",
		},
		{
			duration: "40m3s",
			want:     "0:40:03",
		},
	}

	for i, tt := range tests {
		d, err := time.ParseDuration(tt.duration)
		if err != nil {
			t.Fatalf("[%d]: could not parse duration %v: %v", i, tt.want, err)
		}

		got := FormatDuration(d)
		if got != tt.want {
			t.Errorf("[%d]: got %v, wanted %v", i, got, tt.want)
		}
	}
}
