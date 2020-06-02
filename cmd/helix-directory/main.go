// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/ethulhu/helix/flag"
	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/fileserver"

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

	cd, err := fileserver.NewContentDirectory(basePath, fmt.Sprintf("http://%v/objects/", httpConn.Addr()))
	if err != nil {
		panic(fmt.Sprintf("could not create ContentDirectory object: %v", err))
	}

	device.Handle(contentdirectory.Version1, contentdirectory.ServiceID, contentdirectory.SCPD, contentdirectory.SOAPHandler{cd})
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
