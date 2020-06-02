// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Package fileserver is a basic "serve a directory" ContentDirectory handler.
package fileserver

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"

	log "github.com/sirupsen/logrus"
)

type (
	ContentDirectory struct {
		basePath string
		baseURL  *url.URL
	}
)

func NewContentDirectory(basePath, baseURL string) (*ContentDirectory, error) {
	maybeURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse base URL: %w", err)
	}

	return &ContentDirectory{
		basePath: basePath,
		baseURL:  maybeURL,
	}, nil
}

func (cd *ContentDirectory) BrowseMetadata(_ context.Context, id upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	fields := log.Fields{
		"method": "BrowseMetadata",
		"object": id,
	}

	p, ok := cd.path(id)
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

	item, ok, err := cd.itemFromPath(p)
	if err != nil {
		fields["error"] = err
		log.WithFields(fields).Warning("could not describe item from path")
		return nil, upnpav.ErrActionFailed
	}
	if !ok {
		log.WithFields(fields).Warning("item exists but is not a media item")
		return nil, contentdirectory.ErrNoSuchObject
	}
	return &upnpav.DIDLLite{Items: []upnpav.Item{item}}, nil
}

func (cd *ContentDirectory) path(id upnpav.ObjectID) (string, bool) {
	if id == contentdirectory.Root {
		return cd.basePath, true
	}

	maybePath := path.Clean(path.Join(cd.basePath, string(id)))
	if strings.HasPrefix(maybePath, ".") || strings.HasPrefix(maybePath, "/") {
		return "", false
	}
	return maybePath, true
}

func (cd *ContentDirectory) containerFromPath(p string) (upnpav.Container, error) {
	container := upnpav.Container{
		Object: upnpav.Object{
			ID:     upnpav.ObjectID(p),
			Class:  upnpav.StorageFolder,
			Parent: upnpav.ObjectID(path.Dir(p)), // TODO: wat.
			// Parent: upnpav.ObjectID("-1"),
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

func (cd *ContentDirectory) itemFromPath(p string) (upnpav.Item, bool, error) {
	var class upnpav.Class
	mimetype := mime.TypeByExtension(path.Ext(p))
	switch strings.Split(mimetype, "/")[0] {
	case "audio":
		class = upnpav.AudioItem
	case "video":
		class = upnpav.VideoItem
	default:
		return upnpav.Item{}, false, nil
	}

	uri := *(cd.baseURL)
	uri.Path = path.Join(uri.Path, p)

	item := upnpav.Item{
		Object: upnpav.Object{
			ID:    upnpav.ObjectID(p),
			Class: class,
			Title: path.Base(p),
		},
		Resources: []upnpav.Resource{{
			URI: (&uri).String(),
			ProtocolInfo: &upnpav.ProtocolInfo{
				Protocol:      upnpav.ProtocolHTTP,
				ContentFormat: mimetype,
			},
		}},
	}

	// TODO: something with ffprobe, probably.

	return item, true, nil
}
func (cd *ContentDirectory) BrowseChildren(_ context.Context, parent upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	fields := log.Fields{
		"method": "BrowseChildren",
		"object": parent,
	}

	p, ok := cd.path(parent)
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

	for _, fi := range fs {
		if fi.IsDir() {
			container, err := cd.containerFromPath(path.Join(p, fi.Name()))
			if err != nil {
				fields["error"] = err
				log.WithFields(fields).Warning("could not create container from path")
				continue
			}
			didllite.Containers = append(didllite.Containers, container)
		} else {
			item, ok, err := cd.itemFromPath(path.Join(p, fi.Name()))
			if err != nil {
				fields["error"] = err
				log.WithFields(fields).Warning("could not create item from path")
				continue
			}
			if !ok {
				continue
			}
			didllite.Items = append(didllite.Items, item)
		}
	}
	return didllite, nil
}
func (cd *ContentDirectory) SearchCapabilities(_ context.Context) ([]string, error) {
	return []string{"dc:title"}, nil
}
func (cd *ContentDirectory) SortCapabilities(_ context.Context) ([]string, error) {
	return nil, nil
}
func (cd *ContentDirectory) SystemUpdateID(_ context.Context) (uint, error) {
	return 0, nil
}
func (cd *ContentDirectory) Search(_ context.Context, _ upnpav.ObjectID, _ search.Criteria) (*upnpav.DIDLLite, error) {
	return nil, nil
}