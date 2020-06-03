// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package media

import (
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
)

func TestCoverArtForPath(t *testing.T) {
	tests := []struct {
		fs    fakeFS
		paths []string
		want  [][]string
	}{
		{
			fs: fakeFS{
				"/music":            true,
				"/music/folder.jpg": false,
			},
			paths: []string{"/music"},
			want: [][]string{
				{"/music/folder.jpg"},
			},
		},
		{
			fs: fakeFS{
				"/music": true,
			},
			paths: []string{"/music"},
			want:  [][]string{nil},
		},
		{
			fs: fakeFS{
				"/music":            true,
				"/music/folder.jpg": false,
			},
			paths: []string{"/music"},
			want: [][]string{
				{"/music/folder.jpg"},
			},
		},
		{
			fs: fakeFS{
				"/music":            true,
				"/music/folder.jpg": false,
				"/music/foo.mp3":    false,
			},
			paths: []string{"/music/foo.mp3"},
			want: [][]string{
				{"/music/folder.jpg"},
			},
		},
		{
			fs: fakeFS{
				"/music":             true,
				"/music/folder.jpg":  false,
				"/music/foo.mp3":     false,
				"/music/foo.png":     false,
				"/music/foo.mp3.jpg": false,
			},
			paths: []string{"/music/foo.mp3"},
			want: [][]string{
				{"/music/foo.mp3.jpg", "/music/foo.png"},
			},
		},
		{
			fs: fakeFS{
				"/music":             true,
				"/music/folder.jpg":  false,
				"/music/foo.mp3":     false,
				"/music/foo.mp3.jpg": false,
				"/music/foo 2.mp3":   false,
				"/music/foo 2.png":   false,
			},
			paths: []string{"/music/foo.mp3"},
			want: [][]string{
				{"/music/foo.mp3.jpg"},
			},
		},
	}

	for i, tt := range tests {
		got := coverArtForPaths(tt.fs, tt.paths)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]: coverArtForPath(_, %q) == %q, want %q", i, tt.paths, got, tt.want)
		}
	}
}

// true == directory, false == file
type fakeFS map[string]bool

func (fs fakeFS) Stat(p string) (os.FileInfo, error) {
	p = path.Clean(p)
	if isDir, ok := fs[p]; ok {
		return fakeFileInfo{name: path.Base(p), isDir: isDir}, nil
	}
	return nil, os.ErrNotExist
}
func (fs fakeFS) List(p string) ([]os.FileInfo, error) {
	p = path.Clean(p)
	isDir, ok := fs[p]
	if !ok {
		return nil, os.ErrNotExist
	}
	if !isDir {
		return nil, os.ErrInvalid
	}

	var fileInfos []os.FileInfo
	for f, isDir := range fs {
		if p == path.Dir(f) {
			fileInfos = append(fileInfos, fakeFileInfo{name: path.Base(f), isDir: isDir})
		}
	}
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].Name() < fileInfos[j].Name()
	})
	return fileInfos, nil
}

type fakeFileInfo struct {
	os.FileInfo
	name  string
	isDir bool
}

func (ffi fakeFileInfo) IsDir() bool {
	return ffi.isDir
}
func (ffi fakeFileInfo) Name() string {
	return ffi.name
}
