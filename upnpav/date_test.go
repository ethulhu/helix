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

func TestParseDate(t *testing.T) {
	tests := []struct {
		raw     string
		want    Date
		wantErr bool
	}{
		{
			raw:  "1984-04-01",
			want: Date{time.Date(1984, 4, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			raw:  "1984-04-01T12:24:05",
			want: Date{time.Date(1984, 4, 1, 12, 24, 5, 0, time.UTC)},
		},
	}

	for i, tt := range tests {
		got, err := ParseDate(tt.raw)
		if !tt.wantErr && err != nil {
			t.Errorf("[%d]: got error: %v", i, err)
		}
		if tt.wantErr && err == nil {
			t.Errorf("[%d]: expected error", i)
		}
		if got.Time != tt.want.Time {
			t.Errorf("[%d]: got %v, want %v", i, got, tt.want)
		}
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		date Date
		want string
	}{
		{
			date: Date{time.Date(1984, 4, 1, 0, 0, 0, 0, time.UTC)},
			want: "1984-04-01",
		},
		{
			date: Date{time.Date(1984, 4, 1, 12, 24, 0, 0, time.UTC)},
			want: "1984-04-01",
		},
	}

	for i, tt := range tests {
		got := tt.date.String()
		if got != tt.want {
			t.Errorf("[%d]: got %v, want %v", i, got, tt.want)
		}
	}
}

func TestDateMarshalXML(t *testing.T) {
	tests := []struct {
		data dateTestDocument
		want string
	}{
		{
			data: dateTestDocument{
				Date: Date{time.Date(1984, 4, 1, 0, 0, 0, 0, time.UTC)},
			},
			want: `<test><date>1984-04-01</date></test>`,
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

func TestDateUnmarshalXML(t *testing.T) {
	tests := []struct {
		raw  string
		want dateTestDocument
	}{
		{
			raw: `<test><date>1984-04-01</date></test>`,
			want: dateTestDocument{
				Date: Date{time.Date(1984, 4, 1, 0, 0, 0, 0, time.UTC)},
			},
		},
		{
			raw: `<test><date>1984-04-01T12:24:05</date></test>`,
			want: dateTestDocument{
				Date: Date{time.Date(1984, 4, 1, 12, 24, 5, 0, time.UTC)},
			},
		},
	}

	for i, tt := range tests {
		var got dateTestDocument
		if err := xml.Unmarshal([]byte(tt.raw), &got); err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}
		if !reflect.DeepEqual(got.Date, tt.want.Date) {
			t.Errorf("[%d]: got %v, want %v", i, got.Date, tt.want.Date)
		}
	}
}

type dateTestDocument struct {
	XMLName xml.Name `xml:"test"`
	Date    Date     `xml:"date"`
}
