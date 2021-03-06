// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package connectionmanager

import (
	"context"

	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnp/scpd"
	"github.com/ethulhu/helix/upnpav"
)

type (
	Interface interface {
		// ProtocolInfo lists the protocols that the device can send and receive, respectively.
		ProtocolInfo(context.Context) ([]upnpav.ProtocolInfo, []upnpav.ProtocolInfo, error)
	}
)

const (
	Version1  = upnp.URN("urn:schemas-upnp-org:service:ConnectionManager:1")
	Version2  = upnp.URN("urn:schemas-upnp-org:service:ConnectionManager:2")
	ServiceID = upnp.ServiceID("urn:upnp-org:serviceId:ConnectionManager")
)

var SCPD = scpd.Must(scpd.Merge(
	scpd.Must(scpd.FromAction(getProtocolInfo, getProtocolInfoRequest{}, getProtocolInfoResponse{})),
))
