// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/ethulhu/helix/flag"
	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"

	log "github.com/sirupsen/logrus"
)

var (
	basePath = flag.Custom("path", "", "path to serve", func(raw string) (interface{}, error) {
		if raw == "" {
			return nil, errors.New("must not be empty")
		}
		return raw, nil
	})

	iface = flag.Custom("interface", "", "interface to listen on (will try to find a Private IPv4 if unset)", func(raw string) (interface{}, error) {
		if raw == "" {
			return (*net.Interface)(nil), nil
		}
		return net.InterfaceByName(raw)
	})
)

func main() {
	flag.Parse()

	basePath := (*basePath).(string)
	iface := (*iface).(*net.Interface)

	ip, err := suitableIP(iface)
	if err != nil {
		name := "ALL"
		if iface != nil {
			name = iface.Name
		}
		log.WithFields(log.Fields{
			"interface": name,
			"error":     err,
		}).Fatal("could not find suitable serving IP")
	}
	addr := &net.TCPAddr{
		IP: ip,
	}

	httpConn, err := net.Listen("tcp", addr.String())
	if err != nil {
		log.WithFields(log.Fields{
			"listener": addr,
			"error":    err,
		}).Fatal("could not create HTTP listener")
	}
	defer httpConn.Close()
	log.WithFields(log.Fields{
		"listener": httpConn.Addr(),
	}).Info("created HTTP listener")

	device := upnp.NewDevice("Helix", "uuid:941b0ec2-aca8-4ce1-b64a-329f5762864d")
	device.DeviceType = contentdirectory.DeviceType
	device.Manufacturer = "Eth Morgan"
	device.ManufacturerURL = "https://ethulhu.co.uk"
	device.ModelDescription = "Helix"
	device.ModelName = "Helix"
	device.ModelNumber = "42"
	device.ModelURL = "https://ethulhu.co.uk"
	device.SerialNumber = "00000000"

	cd := contentDirectory{
		BasePath: basePath,
		BaseURL: &url.URL{
			Scheme: "http",
			Host:   httpConn.Addr().String(),
			Path:   "/objects/",
		},
	}
	device.Handle(contentdirectory.Version1, contentdirectory.ServiceID, contentdirectory.SCPD, contentdirectory.SOAPHandler{&cd})
	device.Handle(connectionmanager.Version1, connectionmanager.ServiceID, connectionmanager.SCPD, nil)

	mux := http.NewServeMux()
	mux.Handle("/objects/", http.StripPrefix("/objects/", http.FileServer(http.Dir(basePath))))
	mux.Handle("/", device)

	server := &http.Server{Handler: mux}

	go func() {
		log.WithFields(log.Fields{
			"listener": httpConn.Addr(),
		}).Info("serving HTTP")
		if err := server.Serve(httpConn); err != nil {
			log.WithFields(log.Fields{
				"listener": httpConn.Addr(),
				"error":    err,
			}).Fatal("could not serve HTTP")
		}
	}()

	if err := upnp.BroadcastDevice(device, httpConn.Addr(), nil); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("could serve SSDP")
	}
}

type contentDirectory struct {
	BasePath string
	BaseURL  *url.URL
}

func (cd *contentDirectory) BrowseMetadata(_ context.Context, id upnpav.ObjectID) (*upnpav.DIDLLite, error) {
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
func (cd *contentDirectory) BrowseChildren(_ context.Context, parent upnpav.ObjectID) (*upnpav.DIDLLite, error) {
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

func suitableIP(iface *net.Interface) (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if iface != nil {
		addrs, err = iface.Addrs()
	}
	if err != nil {
		return nil, fmt.Errorf("could not list addresses: %w", err)
	}

	err = errors.New("interface has no addresses")
	for _, addr := range addrs {
		addr, ok := addr.(*net.IPNet)
		if !ok {
			err = errors.New("interface has no IP addresses")
			continue
		}
		if addr.IP.To4() == nil {
			err = errors.New("interface has no IPv4 addresses")
			continue
		}
		ip := addr.IP.To4()

		// Default IP must be a "LAN IP".
		// TODO: support 172.16.0.0/12
		if iface == nil && !(ip[0] == 10 || (ip[0] == 192 && ip[1] == 168)) {
			err = errors.New("interface has no Private IPv4 addresses")
			continue
		}

		return ip, nil
	}
	return nil, err
}

func (cd *contentDirectory) path(id upnpav.ObjectID) (string, bool) {
	if id == contentdirectory.Root {
		return cd.BasePath, true
	}

	maybePath := path.Clean(path.Join(cd.BasePath, string(id)))
	if strings.HasPrefix(maybePath, ".") || strings.HasPrefix(maybePath, "/") {
		return "", false
	}
	return maybePath, true
}

func (cd *contentDirectory) containerFromPath(p string) (upnpav.Container, error) {
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

func (cd *contentDirectory) itemFromPath(p string) (upnpav.Item, bool, error) {
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

	uri := *(cd.BaseURL)
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
