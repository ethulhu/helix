// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/ethulhu/helix/flag"
	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"

	log "github.com/sirupsen/logrus"
)

var (
	filePath = flag.String("path", "", "path to serve")

	iface = flag.Custom("interface", "", "interface to listen on (will try to find a Private IPv4 if unset)", func(raw string) (interface{}, error) {
		if raw == "" {
			return (*net.Interface)(nil), nil
		}
		return net.InterfaceByName(raw)
	})
)

func main() {
	flag.Parse()

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

	d := upnp.NewDevice("Helix", "uuid:941b0ec2-aca8-4ce1-b64a-329f5762864d")
	d.DeviceType = contentdirectory.DeviceType
	d.Manufacturer = "Eth Morgan"
	d.ManufacturerURL = "https://ethulhu.co.uk"
	d.ModelDescription = "Helix"
	d.ModelName = "Helix"
	d.ModelNumber = "42"
	d.ModelURL = "https://ethulhu.co.uk"
	d.SerialNumber = "00000000"

	d.Handle(contentdirectory.Version1, contentdirectory.ServiceID, contentdirectory.SCPD, contentdirectory.SOAPHandler{&contentDirectory{}})
	d.Handle(connectionmanager.Version1, connectionmanager.ServiceID, connectionmanager.SCPD, nil)

	go func() {
		server := &http.Server{Handler: d}
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

	if err := upnp.BroadcastDevice(d, httpConn.Addr(), nil); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("could serve SSDP")
	}
}

type contentDirectory struct {
}

func (cd *contentDirectory) BrowseMetadata(_ context.Context, id upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	if id == contentdirectory.Root {
		id = upnpav.ObjectID("")
	}
	p := path.Join(*filePath, string(id))

	fi, err := os.Stat(p)
	if err != nil {
		log.Printf("could not open %v: %v", p, err)
		return nil, err
	}
	didllite := &upnpav.DIDLLite{
		Containers: []upnpav.Container{
			{
				Object: upnpav.Object{
					ID:     id,
					Title:  fi.Name(),
					Date:   &upnpav.Date{fi.ModTime()},
					Class:  upnpav.StorageFolder,
					Parent: upnpav.ObjectID("-1"),
				},
			},
		},
	}

	fs, err := ioutil.ReadDir(p)
	if err != nil {
		log.Printf("could not open %v: %v", p, err)
		return nil, err
	}
	didllite.Containers[0].ChildCount = len(fs)

	return didllite, nil
}
func (cd *contentDirectory) BrowseChildren(_ context.Context, parent upnpav.ObjectID) (*upnpav.DIDLLite, error) {
	if parent == contentdirectory.Root {
		parent = upnpav.ObjectID("")
	}
	p := path.Join(*filePath, string(parent))
	fs, err := ioutil.ReadDir(p)
	if err != nil {
		log.Printf("could not open %v: %v", p, err)
		return nil, err
	}
	didllite := &upnpav.DIDLLite{}
	for _, fi := range fs {
		didllite.Containers = append(didllite.Containers, upnpav.Container{
			Object: upnpav.Object{
				ID:     upnpav.ObjectID(path.Join(*filePath, string(parent), fi.Name())),
				Date:   &upnpav.Date{fi.ModTime()},
				Title:  fi.Name(),
				Class:  upnpav.StorageFolder,
				Parent: parent,
			},
		})
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
	return &upnpav.DIDLLite{
		Items: []upnpav.Item{
			{
				Object: upnpav.Object{
					Title: "hello",
					Class: upnpav.AudioItem,
				},
			},
		},
	}, nil
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
