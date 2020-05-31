// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package avtransport

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnpav"
)

type (
	client struct{ soap.Interface }
)

func NewClient(soapClient soap.Interface) Interface {
	return &client{soapClient}
}

func (c *client) call(ctx context.Context, method string, input, output interface{}) error {
	req, err := xml.Marshal(input)
	if err != nil {
		panic(fmt.Sprintf("could not marshal SOAP request: %v", err))
	}

	rsp, err := c.Call(ctx, string(Version1), method, req)
	if err != nil {
		return err
	}
	return xml.Unmarshal(rsp, output)
}

func (c *client) Play(ctx context.Context) error {
	req := playRequest{
		InstanceID: 0,
		Speed:      "1",
	}
	return c.call(ctx, play, req, nil)
}
func (c *client) Pause(ctx context.Context) error {
	req := pauseRequest{InstanceID: 0}
	return c.call(ctx, pause, req, nil)
}
func (c *client) Next(ctx context.Context) error {
	req := nextRequest{InstanceID: 0}
	return c.call(ctx, next, req, nil)
}
func (c *client) Previous(ctx context.Context) error {
	req := previousRequest{InstanceID: 0}
	return c.call(ctx, previous, req, nil)
}
func (c *client) Stop(ctx context.Context) error {
	req := stopRequest{InstanceID: 0}
	return c.call(ctx, stop, req, nil)
}
func (c *client) Seek(ctx context.Context, d time.Duration) error {
	req := seekRequest{
		Unit:   SeekRelativeTime,
		Target: upnpav.FormatDuration(d),
	}
	return c.call(ctx, seek, req, nil)
}

func (c *client) MediaInfo(ctx context.Context) (string, *upnpav.DIDLLite, string, *upnpav.DIDLLite, error) {
	req := getMediaInfoRequest{InstanceID: 0}
	rsp := getMediaInfoResponse{}
	if err := c.call(ctx, getMediaInfo, req, &rsp); err != nil {
		return "", nil, "", nil, err
	}

	var currentDIDL *upnpav.DIDLLite
	if len(rsp.CurrentMetadata) != 0 {
		currentDIDL = &upnpav.DIDLLite{}
		if err := xml.Unmarshal(rsp.CurrentMetadata, currentDIDL); err != nil {
			return rsp.CurrentURI, nil, rsp.NextURI, nil, fmt.Errorf("could not unmarshal metadata: %w", err)
		}
	}

	var nextDIDL *upnpav.DIDLLite
	if len(rsp.CurrentMetadata) != 0 {
		nextDIDL = &upnpav.DIDLLite{}
		if err := xml.Unmarshal(rsp.CurrentMetadata, nextDIDL); err != nil {
			return rsp.CurrentURI, currentDIDL, rsp.NextURI, nil, fmt.Errorf("could not unmarshal metadata: %w", err)
		}
	}
	return rsp.CurrentURI, currentDIDL, rsp.NextURI, nextDIDL, nil
}
func (c *client) PositionInfo(ctx context.Context) (string, *upnpav.DIDLLite, time.Duration, time.Duration, error) {
	req := getPositionInfoRequest{InstanceID: 0}
	rsp := getPositionInfoResponse{}
	if err := c.call(ctx, getPositionInfo, req, &rsp); err != nil {
		return "", nil, 0, 0, err
	}

	var didl *upnpav.DIDLLite
	if len(rsp.Metadata) != 0 {
		didl = &upnpav.DIDLLite{}
		if err := xml.Unmarshal(rsp.Metadata, didl); err != nil {
			return rsp.URI, nil, 0, 0, fmt.Errorf("could not unmarshal metadata: %w", err)
		}
	}
	duration, err := upnpav.ParseDuration(rsp.Duration)
	if err != nil && rsp.Duration != "" {
		return rsp.URI, didl, 0, 0, fmt.Errorf("could not unmarshal duration %q: %w", rsp.Duration, err)
	}
	progress, err := upnpav.ParseDuration(rsp.RelativeTime)
	if err != nil && rsp.RelativeTime != "" {
		return rsp.URI, didl, duration, 0, fmt.Errorf("could not unmarshal partial time %q: %w", rsp.RelativeTime, err)
	}

	return rsp.URI, didl, duration, progress, nil
}
func (c *client) TransportInfo(ctx context.Context) (State, Status, error) {
	req := getTransportInfoRequest{}
	rsp := getTransportInfoResponse{}
	if err := c.call(ctx, getTransportInfo, req, &rsp); err != nil {
		return State(""), Status(""), err
	}
	if rsp == (getTransportInfoResponse{}) {
		return State(""), Status(""), errors.New("received an empty GetTransportInfoResponse")
	}
	return rsp.State, rsp.Status, nil
}

func (c *client) SetCurrentURI(ctx context.Context, uri string, metadata *upnpav.DIDLLite) error {
	metadataBytes, err := marshalAndMaybeCreateDIDL(uri, metadata)
	if err != nil {
		return err
	}

	req := setAVTransportURIRequest{
		InstanceID:      0,
		CurrentURI:      uri,
		CurrentMetadata: metadataBytes,
	}
	return c.call(ctx, setAVTransportURI, req, nil)
}
func (c *client) SetNextURI(ctx context.Context, uri string, metadata *upnpav.DIDLLite) error {
	metadataBytes, err := marshalAndMaybeCreateDIDL(uri, metadata)
	if err != nil {
		return err
	}

	req := setNextAVTransportURIRequest{
		InstanceID:   0,
		NextURI:      uri,
		NextMetadata: metadataBytes,
	}
	return c.call(ctx, setNextAVTransportURI, req, nil)
}
func marshalAndMaybeCreateDIDL(uri string, metadata *upnpav.DIDLLite) ([]byte, error) {
	if metadata == nil {
		var err error
		metadata, err = upnpav.DIDLForURI(uri)
		if err != nil {
			return nil, fmt.Errorf("could not create DIDL-Lite for URI %v: %w", uri, err)
		}
	}
	return []byte(metadata.String()), nil
}
