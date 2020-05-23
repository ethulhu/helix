// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"encoding/xml"
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

type (
	// ProtocolInfo is a UPnP AV ProtocolInfo string.
	ProtocolInfo struct {
		Protocol Protocol
		// Network should be "*" for http-get and rtsp-rtp-udp, but can have other values for other
		Network string
		// ContentFormat should be the MIME-type for http-get, or the RTP payload type for rtsp-rtp-udp.
		ContentFormat string
		// AdditionalInfo is frequently "*", but can be used by some formats, e.g. DLNA.ORG_PN extensions.
		AdditionalInfo string
	}

	Protocol string
)

const (
	ProtocolHTTP = Protocol("http-get")
	ProtocolRTSP = Protocol("rtsp-rtp-udp")
)

var (
	ErrUnknownMIMEType = errors.New("could not find valid MIME-type for URI")
)

func ParseProtocolInfo(raw string) (*ProtocolInfo, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("ProtocolInfo must have 4 parts, found %v", len(parts))
	}
	return &ProtocolInfo{
		Protocol:       Protocol(parts[0]),
		Network:        parts[1],
		ContentFormat:  parts[2],
		AdditionalInfo: parts[3],
	}, nil
}

func ProtocolInfoForURI(uri string) (*ProtocolInfo, error) {
	mimeType := mime.TypeByExtension(filepath.Ext(uri))
	if mimeType == "" {
		return nil, ErrUnknownMIMEType
	}
	return &ProtocolInfo{
		Protocol:       ProtocolHTTP,
		Network:        "*",
		ContentFormat:  mimeType,
		AdditionalInfo: "*",
	}, nil
}

func (p *ProtocolInfo) String() string {
	network := "*"
	if p.Network != "" {
		network = p.Network
	}

	additionalInfo := "*"
	if p.AdditionalInfo != "" {
		additionalInfo = p.AdditionalInfo
	}

	return fmt.Sprintf("%s:%s:%s:%s", p.Protocol, network, p.ContentFormat, additionalInfo)
}

func (p *ProtocolInfo) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Name:  name,
		Value: p.String(),
	}, nil
}
func (p *ProtocolInfo) UnmarshalXMLAttr(attr xml.Attr) error {
	pi, err := ParseProtocolInfo(attr.Value)
	if err != nil {
		return err
	}
	p.Protocol = pi.Protocol
	p.Network = pi.Network
	p.ContentFormat = pi.ContentFormat
	p.AdditionalInfo = pi.AdditionalInfo
	return nil
}
