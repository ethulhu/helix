// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Package jackalope is a Jackalope-enhanced ContentDirectory handler.
package jackalope

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethulhu/helix/logger"
	"github.com/ethulhu/helix/media"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
	"go.eth.moe/jackalope"
	"go.eth.moe/jackalope/query"
)

type (
	contentDirectory struct {
		basePath string
		baseURL  *url.URL

		metadataCache media.MetadataCache

		jackalope jackalope.Interface
	}
)

func NewContentDirectory(basePath, baseURL string, metadataCache media.MetadataCache, jackalope jackalope.Interface) (contentdirectory.Interface, error) {
	maybeURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse base URL: %w", err)
	}

	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path: %w", err)
	}

	go func() {
		log := logger.Background()
		log.AddField("path", absPath)
		log.Info("warming metadata cache")

		start := time.Now()
		metadataCache.Warm(absPath)

		log.AddField("duration", time.Since(start))
		log.Info("finished warming metadata cache")
	}()

	return &contentDirectory{
		basePath:      absPath,
		baseURL:       maybeURL,
		metadataCache: metadataCache,
		jackalope:     jackalope,
	}, nil
}

func (cd *contentDirectory) BrowseMetadata(ctx context.Context, id upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	log, ctx := logger.FromContext(ctx)
	log.AddField("jackalope.method", "BrowseMetadata")
	log.AddField("object", id)

	if id == contentdirectory.Root {
		return &upnpav.DIDLLite{
			Containers: []upnpav.Container{{
				Object: upnpav.Object{
					ID:     id,
					Title:  path.Base(cd.basePath),
					Parent: upnpav.ObjectID("-1"),
					Class:  upnpav.StorageFolder,
				},
			}},
		}, nil
	}

	if query, ok := queryForObjectID(id); ok {
		container, err := cd.containerForQuery(id, query)
		if err != nil {
			log.AddField("jackalope.query", query)
			log.WithError(err).Error("could not describe container for query")
			return nil, upnpav.ErrActionFailed
		}
		return &upnpav.DIDLLite{Containers: []upnpav.Container{container}}, nil
	}

	p, ok := pathForObjectID(cd.basePath, id)
	if !ok {
		log.Warning("bad path")
		return nil, contentdirectory.ErrNoSuchObject
	}

	fi, err := os.Stat(p)
	if errors.Is(err, os.ErrNotExist) {
		log.Warning("path does not exist")
		return nil, contentdirectory.ErrNoSuchObject
	}
	if err != nil {
		log.WithError(err).Error("could not stat path")
		return nil, upnpav.ErrActionFailed
	}
	if fi.IsDir() {
		log.Warning("path is directory")
		return nil, contentdirectory.ErrNoSuchObject
	}
	if !media.IsAudioOrVideo(p) {
		log.Warning("item exists but is not a media item")
		return nil, contentdirectory.ErrNoSuchObject
	}

	items, err := cd.itemsForPaths(p)
	if err != nil {
		log.WithError(err).Warning("could not describe item from path")
		return nil, upnpav.ErrActionFailed
	}
	return &upnpav.DIDLLite{Items: items}, nil
}

func (cd *contentDirectory) BrowseChildren(ctx context.Context, id upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	log, ctx := logger.FromContext(ctx)
	log.AddField("jackalope.method", "BrowseChildren")
	log.AddField("object", id)

	if id == contentdirectory.Root {
		containers, err := cd.containersForPaths(id)
		if err != nil {
			log.WithError(err).Error("could not list tags from Jackalope")
			return nil, upnpav.ErrActionFailed
		}
		return &upnpav.DIDLLite{Containers: containers}, nil
	}

	query, ok := queryForObjectID(id)
	if !ok {
		log.Warning("bad query")
		return nil, contentdirectory.ErrNoSuchObject
	}

	paths, err := cd.jackalope.Query(query)
	if err != nil {
		log.WithError(err).Error("could not query Jackalope")
		return nil, upnpav.ErrActionFailed
	}
	paths = filterOutNonExistant(paths)

	containers, err := cd.containersForPaths(id, paths...)
	if err != nil {
		log.WithError(err).Error("could not list tags from Jackalope")
		return nil, upnpav.ErrActionFailed
	}

	items, err := cd.itemsForPaths(paths...)
	if err != nil {
		log.WithError(err).Warning("could not describe items from path")
		return nil, upnpav.ErrActionFailed
	}

	return &upnpav.DIDLLite{Containers: containers, Items: items}, nil
}

func (cd *contentDirectory) SearchCapabilities(_ context.Context) ([]string, error) {
	return nil, nil
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

func queryForObjectID(id upnpav.ObjectID) (string, bool) {
	if _, err := query.Parse(string(id)); err == nil {
		return string(id), true
	}
	return "", false
}

func (cd *contentDirectory) containersForPaths(parent upnpav.ObjectID, paths ...string) ([]upnpav.Container, error) {
	tags, err := cd.jackalope.Tags(paths...)
	if err != nil {
		return nil, err
	}

	var containers []upnpav.Container
	for _, tag := range tags {
		container, err := cd.containerForQuery(parent, tag)
		if err != nil {
			return nil, fmt.Errorf("could not describe container for tag %q", tag)
		}
		containers = append(containers, container)
	}
	return containers, nil
}
func (cd *contentDirectory) containerForQuery(parent upnpav.ObjectID, query string) (upnpav.Container, error) {
	// TODO: actually get ChildCount.
	return upnpav.Container{
		Object: upnpav.Object{
			ID:     upnpav.ObjectID(query),
			Title:  query,
			Parent: parent,
			Class:  upnpav.StorageFolder,
		},
	}, nil
}

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
				Parent: contentdirectory.Root,
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
func objectIDForPath(basePath, p string) upnpav.ObjectID {
	if relPath, err := filepath.Rel(basePath, p); err == nil && relPath != "." {
		return upnpav.ObjectID(relPath)
	}
	return contentdirectory.Root
}

func filterOutNonExistant(paths []string) []string {
	var out []string
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		out = append(out, p)
	}
	return out
}
