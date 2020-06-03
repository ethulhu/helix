// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package media

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var coverArtNames = map[string]bool{
	"cover":  true,
	"folder": true,
	"thumb":  true,
}

func CoverArtForPaths(paths []string) [][]string {
	return coverArtForPaths(realFS{}, paths)
}

func coverArtForPaths(filesystem fs, paths []string) [][]string {
	listings := map[string][]string{}
	for _, p := range paths {
		p := path.Clean(p)

		fi, err := filesystem.Stat(p)
		if err != nil {
			continue
		}

		dir := p
		if !fi.IsDir() {
			dir = path.Dir(p)
		}

		if _, ok := listings[dir]; ok {
			continue
		}

		fis, err := filesystem.List(dir)
		if err != nil {
			continue
		}
		for _, fi := range fis {
			if !fi.IsDir() && IsImage(fi.Name()) {
				listings[dir] = append(listings[dir], fi.Name())
			}
		}
	}

	var allArtPaths [][]string
	for _, p := range paths {
		p := path.Clean(p)

		// Only directories are in the listings, so if it's not it must be a file.
		if _, ok := listings[p]; !ok {
			dir := path.Dir(p)
			file := path.Base(p)

			artPaths := coverArtForFile(dir, file, listings[dir])

			if len(artPaths) > 0 {
				allArtPaths = append(allArtPaths, artPaths)
				continue
			}

			// Fallback to the parent directory's art.
			p = dir
		}

		allArtPaths = append(allArtPaths, coverArtForDir(p, listings[p]))
	}

	return allArtPaths
}

func coverArtForFile(dir, file string, candidates []string) []string {
	withoutExt := strings.TrimSuffix(file, path.Ext(file))

	var artPaths []string
	for _, candidate := range candidates {
		imageWithoutExt := strings.TrimSuffix(candidate, path.Ext(candidate))
		if imageWithoutExt == file || imageWithoutExt == withoutExt {
			artPaths = append(artPaths, path.Join(dir, candidate))
		}
	}
	return artPaths
}
func coverArtForDir(dir string, candidates []string) []string {
	var artPaths []string
	for _, candidate := range candidates {
		imageWithoutExt := strings.TrimSuffix(candidate, path.Ext(candidate))
		if coverArtNames[imageWithoutExt] {
			artPaths = append(artPaths, path.Join(dir, candidate))
		}
	}
	return artPaths
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
