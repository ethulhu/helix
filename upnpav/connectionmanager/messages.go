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
		Sources commaSeparatedProtocolInfos `xml:"Source" scpd:"SourceProtocolInfo,string"`
		Sinks   commaSeparatedProtocolInfos `xml:"Sink"   scpd:"SinkProtocolInfo,string"`
	}

	prepareForConnectionRequest struct {
		XMLName               xml.Name  `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 PrepareForConnection"`
		RemoteProtocolInfo    string    `xml:"RemoteProtocolInfo"    scpd:"A_ARG_TYPE_ProtocolInfo,string"`
		PeerConnectionManager string    `xml:"PeerConnectionManager" scpd:"A_ARG_TYPE_ConnectionManager,string"`
		PeerConnectionID      int       `xml:"PeerConnectionID"      scpd:"A_ARG_TYPE_ConnectionID,i4"`
		Direction             direction `xml:"Direction"             scpd:"A_ARG_TYPE_Direction,string,Input|Output"`
	}
	prepareForConnectionResponse struct {
		XMLName       xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 PrepareForConnectionResponse"`
		ConnectionID  int      `xml:"ConnectionID"  scpd:"A_ARG_TYPE_ConnectionID,i4"`
		AVTransportID int      `xml:"AVTransportID" scpd:"A_ARG_TYPE_AVTransportID,i4"`
		ResID         int      `xml:"ResID"         scpd:"A_ARG_TYPE_ResID,i4"`
	}

	connectionCompleteRequest struct {
		XMLName      xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 ConnectionComplete"`
		ConnectionID int      `xml:"ConnectionID" scpd:"A_ARG_TYPE_ConnectionID,i4"`
	}
	connectionCompleteResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 ConnectionCompleteResponse"`
	}

	getCurrentConnectionIDsRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetCurrentConnectionIDs"`
	}
	getCurrentConnectionIDsResponse struct {
		XMLName       xml.Name           `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetCurrentConnectionIDsResponse"`
		ConnectionIDs commaSeparatedInts `xml:"ConnectionIDs" scpd:"CurrentConnectionIDs,string"`
	}

	getCurrentConnectionInfoRequest struct {
		XMLName      xml.Name `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetCurrentConnectionInfo"`
		ConnectionID int      `xml:"ConnectionID" scpd:"A_ARG_TYPE_ConnectionID,i4"`
	}
	getCurrentConnectionInfoResponse struct {
		XMLName               xml.Name  `xml:"urn:schemas-upnp-org:service:ConnectionManager:1 GetCurrentConnectionInfoResponse"`
		AVTransportID         int       `xml:"AVTransportID"         scpd:"A_ARG_TYPE_AVTransportID,i4"`
		ResID                 int       `xml:"ResID"                 scpd:"A_ARG_TYPE_ResID,i4"`
		ProtocolInfo          string    `xml:"ProtocolInfo"          scpd:"A_ARG_TYPE_ProtocolInfo,string"`
		PeerConnecitonManager string    `xml:"PeerConnecitonManager" scpd:"A_ARG_TYPE_ConnectionManager,string"`
		PeerConnecitonID      int       `xml:"PeerConnecitonID"      scpd:"A_ARG_TYPE_ConnectionID,i4"`
		Direction             direction `xml:"Direction"             scpd:"A_ARG_TYPE_Direction,string,Input|Output"`
		Status                status    `xml:"Status"                scpd:"A_ARG_TYPE_ConnectionStatus,string,OK|ContentFormatMismatch|InsufficientBandwidth|UnreliableChannel|Unknown"`
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
