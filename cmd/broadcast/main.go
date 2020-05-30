// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"log"
	"net"
	"net/http"

	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnp/scpd"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

var (
	contentDirectorySCPD = scpd.Document{
		SpecVersion: scpd.SpecVersion{
			Major: 1,
			Minor: 1,
		},
	}
)

func main() {
	httpConn, err := net.Listen("tcp", "192.168.69.195:0")
	if err != nil {
		log.Fatalf("could not create HTTP listener: %v", err)
	}
	defer httpConn.Close()

	d := upnp.NewDevice("thingy", "uuid:941b0ec2-aca8-4ce1-b64a-329f5762864d")
	d.HandleURN(contentdirectory.Version1, contentDirectorySCPD, nil)

	httpServer := &http.Server{
		Handler: d,
	}
	go func() {
		if err := httpServer.Serve(httpConn); err != nil {
			log.Fatalf("could not serve HTTP: %v", err)
		}
	}()

	if err := upnp.BroadcastDevice(d, httpConn.Addr(), nil); err != nil {
		log.Fatalf("could not broadcast device: %v", err)
	}
}