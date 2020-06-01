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
	if output != nil {
		return xml.Unmarshal(rsp, output)
	}
	return nil
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
		Target: upnpav.Duration{d}.String(),
	}
	return c.call(ctx, seek, req, nil)
}

func (c *client) MediaInfo(ctx context.Context) (string, *upnpav.DIDLLite, string, *upnpav.DIDLLite, error) {
	req := getMediaInfoRequest{InstanceID: 0}
	rsp := getMediaInfoResponse{}
	if err := c.call(ctx, getMediaInfo, req, &rsp); err != nil {
		return "", nil, "", nil, err
	}
	return rsp.CurrentURI, &rsp.CurrentMetadata.DIDLLite, rsp.NextURI, &rsp.NextMetadata.DIDLLite, nil
}
func (c *client) PositionInfo(ctx context.Context) (string, *upnpav.DIDLLite, time.Duration, time.Duration, error) {
	req := getPositionInfoRequest{InstanceID: 0}
	rsp := getPositionInfoResponse{}
	if err := c.call(ctx, getPositionInfo, req, &rsp); err != nil {
		return "", nil, 0, 0, err
	}
	return rsp.URI, &rsp.Metadata.DIDLLite, rsp.Duration.Duration, rsp.RelativeTime.Duration, nil
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
	if metadata == nil {
		var err error
		metadata, err = upnpav.DIDLForURI(uri)
		if err != nil {
			return fmt.Errorf("could not create DIDL-Lite for URI %v: %w", uri, err)
		}
	}
	req := setAVTransportURIRequest{
		InstanceID:      0,
		CurrentURI:      uri,
		CurrentMetadata: upnpav.EncodedDIDLLite{*metadata},
	}
	return c.call(ctx, setAVTransportURI, req, nil)
}
func (c *client) SetNextURI(ctx context.Context, uri string, metadata *upnpav.DIDLLite) error {
	if metadata == nil {
		var err error
		metadata, err = upnpav.DIDLForURI(uri)
		if err != nil {
			return fmt.Errorf("could not create DIDL-Lite for URI %v: %w", uri, err)
		}
	}
	req := setNextAVTransportURIRequest{
		InstanceID:   0,
		NextURI:      uri,
		NextMetadata: upnpav.EncodedDIDLLite{*metadata},
	}
	return c.call(ctx, setNextAVTransportURI, req, nil)
}
