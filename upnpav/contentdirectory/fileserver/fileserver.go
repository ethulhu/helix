// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Package fileserver is a basic "serve a directory" contentDirectory handler.
package fileserver

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethulhu/helix/media"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"

	log "github.com/sirupsen/logrus"
)

type (
	contentDirectory struct {
		basePath string
		baseURL  *url.URL

		metadataCache media.MetadataCache
	}
)

func NewContentDirectory(basePath, baseURL string, metadataCache media.MetadataCache) (contentdirectory.Interface, error) {
	maybeURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse base URL: %w", err)
	}

	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path: %w", err)
	}

	go func() {
		fields := log.Fields{"path": absPath}
		log.WithFields(fields).Info("warming metadata cache")

		start := time.Now()
		metadataCache.Warm(absPath)
		fields["duration"] = time.Since(start)

		log.WithFields(fields).Info("finished warming metadata cache")
	}()

	return &contentDirectory{
		basePath: absPath,
		baseURL:  maybeURL,

		metadataCache: metadataCache,
	}, nil
}

func (cd *contentDirectory) BrowseMetadata(_ context.Context, id upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	fields := log.Fields{
		"method": "BrowseMetadata",
		"object": id,
	}

	p, ok := pathForObjectID(cd.basePath, id)
	if !ok {
		log.WithFields(fields).Error("bad path")
		return nil, contentdirectory.ErrNoSuchObject
	}

	fi, err := os.Stat(p)
	if errors.Is(err, os.ErrNotExist) {
		log.WithFields(fields).Info("path does not exist")
		return nil, contentdirectory.ErrNoSuchObject
	}
	if err != nil {
		fields["error"] = err
		log.WithFields(fields).Warning("could not stat path")
		return nil, upnpav.ErrActionFailed
	}

	if fi.IsDir() {
		container, err := cd.containerFromPath(p)
		if err != nil {
			fields["error"] = err
			log.WithFields(fields).Warning("could not describe container from path")
			return nil, upnpav.ErrActionFailed
		}
		return &upnpav.DIDLLite{Containers: []upnpav.Container{container}}, nil
	}

	if !media.IsAudioOrVideo(p) {
		log.WithFields(fields).Warning("item exists but is not a media item")
		return nil, contentdirectory.ErrNoSuchObject
	}

	items, err := cd.itemsForPaths(p)
	if err != nil {
		fields["error"] = err
		log.WithFields(fields).Warning("could not describe item from path")
		return nil, upnpav.ErrActionFailed
	}
	return &upnpav.DIDLLite{Items: items}, nil
}

func (cd *contentDirectory) BrowseChildren(_ context.Context, parent upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	fields := log.Fields{
		"method": "BrowseChildren",
		"object": parent,
	}

	p, ok := pathForObjectID(cd.basePath, parent)
	if !ok {
		log.WithFields(fields).Error("bad path")
		return nil, contentdirectory.ErrNoSuchObject
	}

	fi, err := os.Stat(p)
	if errors.Is(err, os.ErrNotExist) {
		log.WithFields(fields).Info("path does not exist")
		return nil, contentdirectory.ErrNoSuchObject
	}
	if err != nil {
		fields["error"] = err
		log.WithFields(fields).Warning("could not stat path")
		return nil, upnpav.ErrActionFailed
	}

	if !fi.IsDir() {
		log.WithFields(fields).Info("not a directory")
		return nil, nil
	}

	didllite := &upnpav.DIDLLite{}

	fs, err := ioutil.ReadDir(p)
	if err != nil {
		fields["error"] = err
		log.WithFields(fields).Error("could not list directory")
		return didllite, upnpav.ErrActionFailed
	}

	var itemPaths []string
	for _, fi := range fs {
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}

		if !fi.IsDir() {
			if media.IsAudioOrVideo(fi.Name()) {
				itemPaths = append(itemPaths, path.Join(p, fi.Name()))
			}
			continue
		}

		container, err := cd.containerFromPath(path.Join(p, fi.Name()))
		if err != nil {
			fields["error"] = err
			log.WithFields(fields).Warning("could not create container from path")
			continue
		}
		didllite.Containers = append(didllite.Containers, container)
	}

	items, err := cd.itemsForPaths(itemPaths...)
	if err != nil {
		fields["error"] = err
		log.WithFields(fields).Warning("could not create items from paths")
	}
	didllite.Items = items

	return didllite, nil
}
func (cd *contentDirectory) SearchCapabilities(_ context.Context) ([]string, error) {
	return []string{"dc:title"}, nil
}
func (cd *contentDirectory) SortCapabilities(_ context.Context) ([]string, error) {
	return nil, nil
}
func (cd *contentDirectory) SystemUpdateID(_ context.Context) (uint, error) {
	return 0, nil
}
func (cd *contentDirectory) Search(_ context.Context, _ upnpav.ObjectID, _ search.Criteria) (*upnpav.DIDLLite, error) {
	return nil, nil
}

func (cd *contentDirectory) containerFromPath(p string) (upnpav.Container, error) {
	container := upnpav.Container{
		Object: upnpav.Object{
			ID:     objectIDForPath(cd.basePath, p),
			Parent: parentIDForPath(cd.basePath, p),
			Class:  upnpav.StorageFolder,
		},
	}

	fi, err := os.Stat(p)
	if err != nil {
		return container, err
	}
	container.Title = fi.Name()
	container.Date = &upnpav.Date{fi.ModTime()}

	fs, err := ioutil.ReadDir(p)
	if err != nil {
		return container, err
	}
	container.ChildCount = len(fs)

	return container, nil
}

func (cd *contentDirectory) itemsForPaths(paths ...string) ([]upnpav.Item, error) {
	coverArts := media.CoverArtForPaths(paths)
	metadatas := cd.metadataCache.MetadataForPaths(paths)

	var items []upnpav.Item
	for i, p := range paths {
		md := metadatas[i]

		class, err := upnpav.ClassForMIMEType(md.MIMEType)
		if err != nil {
			panic(fmt.Sprintf("should only have audio or video MIME-Types, got %q for path %q", md.MIMEType, p))
		}

		var albumArtURIs []string
		for _, artPath := range coverArts[i] {
			albumArtURIs = append(albumArtURIs, cd.uri(artPath))
		}

		items = append(items, upnpav.Item{
			Object: upnpav.Object{
				ID:     objectIDForPath(cd.basePath, p),
				Parent: parentIDForPath(cd.basePath, p),
				Class:  class,
				Title:  md.Title,
			},
			AlbumArtURIs: albumArtURIs,
			Resources: []upnpav.Resource{{
				URI:      cd.uri(p),
				Duration: &upnpav.Duration{md.Duration},
				ProtocolInfo: &upnpav.ProtocolInfo{
					Protocol:      upnpav.ProtocolHTTP,
					ContentFormat: md.MIMEType,
				},
			}},
		})
	}

	return items, nil
}

func (cd *contentDirectory) uri(p string) string {
	uri := *(cd.baseURL)
	relPath, _ := filepath.Rel(cd.basePath, p)
	uri.Path = path.Join(uri.Path, relPath)
	// TODO: figure out what's actually going wrong here.
	return strings.Replace((&uri).String(), "&", "%26", -1)
}
