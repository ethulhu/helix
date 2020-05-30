// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package connectionmanager

import (
	"context"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

type (
	Interface interface {
		// ProtocolInfo lists the protocols that the device can send and receive, respectively.
		ProtocolInfo(context.Context) ([]*upnpav.ProtocolInfo, []*upnpav.ProtocolInfo, error)
	}
)

const (
	Version1 = ssdp.URN("urn:schemas-upnp-org:service:ConnectionManager:1")
	Version2 = ssdp.URN("urn:schemas-upnp-org:service:ConnectionManager:2")
)
