// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/ethulhu/helix/flag"
	"github.com/ethulhu/helix/flags"
	"github.com/ethulhu/helix/logger"
	"github.com/ethulhu/helix/media"
	"github.com/ethulhu/helix/netutil"
	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/fileserver"
)

var (
	basePath = flag.Custom("path", "", "path to serve", func(raw string) (interface{}, error) {
		if raw == "" {
			return nil, errors.New("must not be empty")
		}
		return raw, nil
	})

	udn          = flag.Custom("udn", "", "UDN to broadcast (if unset, will generate one)", flags.UDN)
	friendlyName = flag.Custom("friendly-name", "", "human-readable name to broadcast (if unset, will generate one)", flags.FriendlyName)
	iface        = flag.Custom("interface", "", "interface to listen on (will try to find a Private IPv4 if unset)", flags.NetInterface)

	disableMetadataCache = flag.Bool("disable-metadata-cache", false, "disable the metadata cache")
)

func main() {
	flag.Parse()

	basePath := (*basePath).(string)
	friendlyName := (*friendlyName).(string)
	iface := (*iface).(*net.Interface)
	udn := (*udn).(string)

	log, _ := logger.FromContext(context.Background())

	ip, err := netutil.SuitableIP(iface)
	if err != nil {
		name := "ALL"
		if iface != nil {
			name = iface.Name
		}
		log.AddField("interface", name)
		log.WithError(err).Fatal("could not find suitable serving IP")
	}
	addr := &net.TCPAddr{
		IP: ip,
	}

	httpConn, err := net.Listen("tcp", addr.String())
	if err != nil {
		log.AddField("listener", addr)
		log.WithError(err).Fatal("could not create HTTP listener")
	}
	defer httpConn.Close()

	device := &upnp.Device{
		Name:             friendlyName,
		UDN:              udn,
		DeviceType:       contentdirectory.DeviceType,
		Manufacturer:     "Eth Morgan",
		ManufacturerURL:  "https://ethulhu.co.uk",
		ModelDescription: "Helix",
		ModelName:        "Helix",
		ModelNumber:      "42",
		ModelURL:         "https://ethulhu.co.uk",
		SerialNumber:     "00000000",
	}

	metadataCache := media.NewMetadataCache()
	if *disableMetadataCache {
		metadataCache = media.NoOpCache{}
	}

	cd, err := fileserver.NewContentDirectory(basePath, fmt.Sprintf("http://%v/objects/", httpConn.Addr()), metadataCache)
	if err != nil {
		log.WithError(err).Fatal("could not create ContentDirectory object")
	}

	device.Handle(contentdirectory.Version1, contentdirectory.ServiceID, contentdirectory.SCPD, contentdirectory.SOAPHandler{cd})
	device.Handle(connectionmanager.Version1, connectionmanager.ServiceID, connectionmanager.SCPD, nil)

	mux := http.NewServeMux()
	mux.Handle("/objects/", http.StripPrefix("/objects/", http.FileServer(http.Dir(basePath))))
	mux.Handle("/", device)

	httpServer := &http.Server{Handler: mux}
	go func() {
		log := log.WithField("http.listener", httpConn.Addr())
		log.Info("serving HTTP")
		if err := httpServer.Serve(httpConn); err != nil {
			log.WithError(err).Fatal("could not serve HTTP")
		}
	}()

	if err := upnp.BroadcastDevice(device, httpConn.Addr(), nil); err != nil {
		log.WithError(err).Fatal("could not serve SSDP")
	}
}
