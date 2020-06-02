// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package fileserver

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

// TODO: encode these with base64 or base32 if a directory browser seems
// unhappy; it may be having issues with spaces or "/".

func pathForObjectID(basePath string, id upnpav.ObjectID) (string, bool) {
	if id == contentdirectory.Root {
		return basePath, true
	}

	maybePath := path.Clean(path.Join(basePath, string(id)))
	if !strings.HasPrefix(maybePath, basePath) {
		return "", false
	}
	return maybePath, true
}

func objectIDForPath(basePath, p string) upnpav.ObjectID {
	if relPath, err := filepath.Rel(basePath, p); err == nil && relPath != "." {
		return upnpav.ObjectID(relPath)
	}
	return contentdirectory.Root
}

func parentIDForPath(basePath, p string) upnpav.ObjectID {
	id := objectIDForPath(basePath, p)
	if id == contentdirectory.Root {
		return upnpav.ObjectID("-1")
	}
	return objectIDForPath(basePath, path.Dir(p))
}
