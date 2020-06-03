// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package media

import (
	"io/ioutil"
	"mime"
	"os"
	"path"
	"strings"
)

var basenames = map[string]bool{
	"cover":  true,
	"folder": true,
	"thumb":  true,
}

func CoverArtForPath(p string) []string {
	paths, _ := coverArtForPath(realFS{}, p)
	return paths
}

type (
	fs interface {
		Stat(string) (os.FileInfo, error)
		List(string) ([]os.FileInfo, error)
	}

	realFS struct{}
)

func (_ realFS) Stat(p string) (os.FileInfo, error) {
	return os.Stat(p)
}
func (_ realFS) List(p string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(p)
}

func coverArtForPath(filesystem fs, p string) ([]string, error) {
	fi, err := filesystem.Stat(p)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		// Check our neighbors, e.g. for foo.mp3:
		// - foo.jpg
		// - foo.mp3.jpg
		//
		// TODO: include foo-cover.jpg ?
		fs, err := filesystem.List(path.Dir(p))
		if err != nil {
			return nil, err
		}
		basename := path.Base(p)
		withoutExt := strings.TrimSuffix(basename, path.Ext(basename))

		var paths []string
		for _, f := range fs {
			ext := path.Ext(f.Name())
			if !strings.HasPrefix(mime.TypeByExtension(ext), "image/") {
				continue
			}

			imageWithoutExt := strings.TrimSuffix(f.Name(), ext)
			if imageWithoutExt == basename || imageWithoutExt == withoutExt {
				paths = append(paths, path.Join(path.Dir(p), f.Name()))
			}
		}
		if len(paths) > 0 {
			return paths, nil
		}

		// Fallback to the parent directory's art.
		p = path.Dir(p)
	}

	fis, err := filesystem.List(p)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, fi := range fis {
		ext := path.Ext(fi.Name())
		basename := strings.TrimSuffix(fi.Name(), ext)
		if basenames[basename] && strings.HasPrefix(mime.TypeByExtension(ext), "image/") {
			paths = append(paths, path.Join(p, fi.Name()))
		}
	}
	return paths, nil
}
