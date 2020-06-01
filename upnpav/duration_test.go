// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		raw     string
		want    Duration
		wantErr bool
	}{
		{
			raw:  "2:00:03",
			want: Duration{2*time.Hour + 3*time.Second},
		},
		{
			raw:  "0:40:03",
			want: Duration{40*time.Minute + 3*time.Second},
		},
		{
			raw:  "00:04:03",
			want: Duration{4*time.Minute + 3*time.Second},
		},
		{
			raw:  "04:03",
			want: Duration{4*time.Minute + 3*time.Second},
		},
		{
			raw:  ":04:03",
			want: Duration{4*time.Minute + 3*time.Second},
		},
		{
			raw:  "04:03.123",
			want: Duration{4*time.Minute + 3*time.Second},
		},
		{
			raw:  "04:03.1/3",
			want: Duration{4*time.Minute + 3*time.Second},
		},
		{
			raw:  "1:04:03",
			want: Duration{1*time.Hour + 4*time.Minute + 3*time.Second},
		},
	}

	for i, tt := range tests {
		got, err := ParseDuration(tt.raw)
		if !tt.wantErr && err != nil {
			t.Errorf("[%d]: got error: %v", i, err)
		}
		if tt.wantErr && err == nil {
			t.Errorf("[%d]: expected error", i)
		}
		if got.Duration != tt.want.Duration {
			t.Errorf("[%d]: got %v, want %v", i, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration Duration
		want     string
	}{
		{
			duration: Duration{3*time.Hour + 2*time.Minute + 1*time.Second},
			want:     "3:02:01",
		},
		{
			duration: Duration{2*time.Minute + 1*time.Second},
			want:     "0:02:01",
		},
		{
			duration: Duration{1 * time.Second},
			want:     "0:00:01",
		},
	}

	for i, tt := range tests {
		got := tt.duration.String()
		if got != tt.want {
			t.Errorf("[%d]: got %v, want %v", i, got, tt.want)
		}
	}
}

func TestDurationMarshalXML(t *testing.T) {
	tests := []struct {
		data durationTestDocument
		want string
	}{
		{
			data: durationTestDocument{
				Duration: Duration{3 * time.Hour},
			},
			want: `<test><duration>3:00:00</duration></test>`,
		},
	}

	for i, tt := range tests {
		bytes, err := xml.Marshal(tt.data)
		if err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}
		got := string(bytes)

		if got != tt.want {
			t.Errorf("[%d]: got %v, want %v", i, got, tt.want)
		}
	}
}

func TestDurationUnmarshalXML(t *testing.T) {
	tests := []struct {
		raw  string
		want durationTestDocument
	}{
		{
			raw: `<test><duration>3:00:00</duration></test>`,
			want: durationTestDocument{
				Duration: Duration{3 * time.Hour},
			},
		},
	}

	for i, tt := range tests {
		var got durationTestDocument
		if err := xml.Unmarshal([]byte(tt.raw), &got); err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}
		if !reflect.DeepEqual(got.Duration, tt.want.Duration) {
			t.Errorf("[%d]: got %v, want %v", i, got.Duration, tt.want.Duration)
		}
	}
}

type durationTestDocument struct {
	XMLName  xml.Name `xml:"test"`
	Duration Duration `xml:"duration"`
}
