// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package media

import (
	"reflect"
	"testing"
	"time"
)

func TestMergeFFProbeOutput(t *testing.T) {
	tests := []struct {
		md      *Metadata
		ffprobe ffprobeOutput
		want    *Metadata
	}{
		{
			md:      &Metadata{},
			ffprobe: ffprobeOutput{},
			want: &Metadata{
				Tags: map[string]string{},
			},
		},
		{
			md: &Metadata{},
			ffprobe: ffprobeOutput{
				Format: ffprobeFormat{
					DurationSeconds: "12.4",
				},
			},
			want: &Metadata{
				Duration: 12 * time.Second,
				Tags: map[string]string{},
			},
		},
		{
			md: &Metadata{
				Title: "foo",
			},
			ffprobe: ffprobeOutput{
				Format: ffprobeFormat{
					DurationSeconds: "12.4",
				},
			},
			want: &Metadata{
				Title:    "foo",
				Duration: 12 * time.Second,
				Tags: map[string]string{},
			},
		},
		{
			md: &Metadata{
				Title: "foo",
			},
			ffprobe: ffprobeOutput{
				Format: ffprobeFormat{
					DurationSeconds: "12.4",
					Tags: map[string]string{
						"TiTlE": "bar",
						"Album": "baz",
					},
				},
			},
			want: &Metadata{
				Title:    "bar",
				Duration: 12 * time.Second,
				Tags: map[string]string{
					"Album": "baz",
				},
			},
		},
	}

	for i, tt := range tests {
		got := tt.md
		if err := mergeFFProbeOutput(got, tt.ffprobe); err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}

		if !reflect.DeepEqual(tt.md, tt.want) {
			t.Errorf("[%d]: got %+v, want %+v", i, tt.md, tt.want)
		}
	}
}
