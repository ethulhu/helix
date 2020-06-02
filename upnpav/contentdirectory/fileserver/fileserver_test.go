// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package fileserver

import (
	"testing"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

func TestObjectIDForPath(t *testing.T) {
	tests := []struct {
		basePath string
		path     string
		want     upnpav.ObjectID
	}{
		{
			basePath: "/mnt/media",
			path:     "",
			want:     contentdirectory.Root,
		},
		{
			basePath: "/mnt/media",
			path:     "/mnt/media/foo",
			want:     upnpav.ObjectID("foo"),
		},
	}

	for i, tt := range tests {
		got := objectIDForPath(tt.basePath, tt.path)

		if got != tt.want {
			t.Errorf("[%d]: objectIDForPath(%q, %q) == %q, want %q", i, tt.basePath, tt.path, got, tt.want)
		}
	}
}

func TestParentIDForPath(t *testing.T) {
	tests := []struct {
		basePath string
		path     string
		want     upnpav.ObjectID
	}{
		{
			basePath: "/mnt/media",
			path:     "",
			want:     upnpav.ObjectID("-1"),
		},
		{
			basePath: "/mnt/media",
			path:     "/mnt/media/foo",
			want:     contentdirectory.Root,
		},
		{
			basePath: "/mnt/media",
			path:     "/mnt/media/foo/bar",
			want:     upnpav.ObjectID("foo"),
		},
	}

	for i, tt := range tests {
		got := parentIDForPath(tt.basePath, tt.path)

		if got != tt.want {
			t.Errorf("[%d]: objectIDForPath(%q, %q) == %q, want %q", i, tt.basePath, tt.path, got, tt.want)
		}
	}
}

func TestPathForObjectID(t *testing.T) {
	tests := []struct {
		basePath string
		object   upnpav.ObjectID
		want     string
		wantOK   bool
	}{
		{
			basePath: "/mnt/media",
			object:   contentdirectory.Root,
			want:     "/mnt/media",
			wantOK:   true,
		},
		{
			basePath: "/mnt/media",
			object:   upnpav.ObjectID("foo"),
			want:     "/mnt/media/foo",
			wantOK:   true,
		},
		{
			basePath: "/mnt/media",
			object:   upnpav.ObjectID("/etc/passwd"),
			want:     "/mnt/media/etc/passwd",
			wantOK:   true,
		},
		{
			basePath: "/mnt/media",
			object:   upnpav.ObjectID("../../etc/passwd"),
			wantOK:   false,
		},
	}

	for i, tt := range tests {
		got, ok := pathForObjectID(tt.basePath, tt.object)

		if ok != tt.wantOK {
			t.Errorf("[%d]: pathForObjectID(%q, %q) == %v, want %v", i, tt.basePath, tt.object, ok, tt.wantOK)
		}

		if got != tt.want {
			t.Errorf("[%d]: pathForObjectID(%q, %q) == %q, want %q", i, tt.basePath, tt.object, got, tt.want)
		}
	}
}
