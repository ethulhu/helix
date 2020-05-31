// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package connectionmanager

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/ethulhu/helix/upnpav"
)

type (
	commaSeparatedProtocolInfos []*upnpav.ProtocolInfo
	commaSeparatedInts          []int

	direction string
	status    string

	getProtocolInfoRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetProtocolInfo"`
	}
	getProtocolInfoResponse struct {
		XMLName xml.Name                    `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetProtocolInfoResponse"`
		Sources commaSeparatedProtocolInfos `xml:"Source"`
		Sinks   commaSeparatedProtocolInfos `xml:"Sink"`
	}

	prepareForConnectionRequest struct {
		XMLName               xml.Name  `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 PrepareForConnection"`
		RemoteProtocolInfo    string    `xml:"RemoteProtocolInfo"`
		PeerConnectionManager string    `xml:"PeerConnectionManager"`
		PeerConnectionID      int       `xml:"PeerConnectionID"`
		Direction             direction `xml:"Direction"`
	}
	prepareForConnectionResponse struct {
		XMLName       xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 PrepareForConnectionResponse"`
		ConnectionID  int      `xml:"ConnectionID"`
		AVTransportID int      `xml:"AVTransportID"`
		ResID         int      `xml:"ResID"`
	}

	connectionCompleteRequest struct {
		XMLName      xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 ConnectionComplete"`
		ConnectionID int      `xml:"ConnectionID"`
	}
	connectionCompleteResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 ConnectionCompleteResponse"`
	}

	getCurrentConnectionIDsRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetCurrentConnectionIDs"`
	}
	getCurrentConnectionIDsResponse struct {
		XMLName       xml.Name           `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetCurrentConnectionIDsResponse"`
		ConnectionIDs commaSeparatedInts `xml:"ConnectionIDs"`
	}

	getCurrentConnectionInfoRequest struct {
		XMLName      xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetCurrentConnectionInfo"`
		ConnectionID int      `xml:"ConnectionID"`
	}
	getCurrentConnectionInfoResponse struct {
		XMLName               xml.Name  `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetCurrentConnectionInfoResponse"`
		AVTransportID         int       `xml:"AVTransportID"`
		ResID                 int       `xml:"ResID"`
		ProtocolInfo          string    `xml:"ProtocolInfo"`
		PeerConnecitonManager string    `xml:"PeerConnecitonManager"`
		PeerConnecitonID      int       `xml:"PeerConnecitonID"`
		Direction             direction `xml:"Direction"`
		Status                status    `xml:"Status"`
	}
)

const (
	input  = direction("Input")
	output = direction("Output")
)

const (
	ok                    = status("OK")
	contentFormatMismatch = status("ContentFormatMismatch")
	insufficientBandwidth = status("InsufficientBandwidth")
	unreliableChannel     = status("UnreliableChannel")
	unknown               = status("Unknown")
)

const (
	getProtocolInfo          = "GetProtocolInfo"
	prepareForConnection     = "PrepareForConnection"
	connectionComplete       = "ConnectionComplete"
	getCurrentConnectionIDs  = "GetCurrentConnectionIDs"
	getCurrentConnectionInfo = "GetCurrentConnectionInfo"
)

func (ps commaSeparatedProtocolInfos) MarshalXML(e *xml.Encoder, el xml.StartElement) error {
	var piStrings []string
	for _, p := range ps {
		piStrings = append(piStrings, p.String())
	}
	return e.EncodeElement(strings.Join(piStrings, ","), el)
}

func (ps *commaSeparatedProtocolInfos) UnmarshalXML(d *xml.Decoder, el xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &el); err != nil {
		return err
	}
	if s == "" {
		return nil
	}

	var protocolInfos []*upnpav.ProtocolInfo
	for _, p := range strings.Split(s, ",") {
		protocolInfo, err := upnpav.ParseProtocolInfo(p)
		if err != nil {
			return fmt.Errorf("could not parse ProtocolInfo %q: %v", p, err)
		}
		protocolInfos = append(protocolInfos, protocolInfo)
	}

	*ps = protocolInfos
	return nil
}
