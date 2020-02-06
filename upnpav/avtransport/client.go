package avtransport

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

type (
	client struct {
		name       string
		soapClient soap.Client
	}
)

func Discover(ctx context.Context) ([]Client, []error, error) {
	devices, errs, err := ssdp.DiscoverDevices(ctx, Version1)

	var clients []Client
	for _, device := range devices {
		soapClient, ok := device.Client(Version1)
		if !ok {
			// TODO: expand this.
			errs = append(errs, fmt.Errorf("could not find ConnectionManager client"))
			continue
		}
		clients = append(clients, NewClient(device.Name(), soapClient))
	}
	return clients, errs, err
}

func NewClient(name string, soapClient soap.Client) Client {
	return &client{
		name:       name,
		soapClient: soapClient,
	}
}

func (c *client) Name() string {
	return c.name
}

func (c *client) call(ctx context.Context, method string, input, output interface{}) error {
	return c.soapClient.Call(ctx, string(Version1), method, input, output)
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

func (c *client) MediaInfo(ctx context.Context) (string, *upnpav.DIDL, error) {
	req := getMediaInfoRequest{InstanceID: 0}
	rsp := getMediaInfoResponse{}
	if err := c.call(ctx, getMediaInfo, req, &rsp); err != nil {
		return "", nil, err
	}

	var didl *upnpav.DIDL
	if len(rsp.CurrentMetadata) != 0 {
		didl = &upnpav.DIDL{}
		if err := xml.Unmarshal(rsp.CurrentMetadata, didl); err != nil {
			return rsp.CurrentURI, nil, fmt.Errorf("could not unmarshal metadata: %w", err)
		}
	}
	return rsp.CurrentURI, didl, nil
}
func (c *client) PositionInfo(ctx context.Context) (string, *upnpav.DIDL, time.Duration, time.Duration, error) {
	req := getPositionInfoRequest{InstanceID: 0}
	rsp := getPositionInfoResponse{}
	if err := c.call(ctx, getPositionInfo, req, &rsp); err != nil {
		return "", nil, 0, 0, err
	}

	var didl *upnpav.DIDL
	if len(rsp.Metadata) != 0 {
		didl = &upnpav.DIDL{}
		if err := xml.Unmarshal(rsp.Metadata, didl); err != nil {
			return rsp.URI, nil, 0, 0, fmt.Errorf("could not unmarshal metadata: %w", err)
		}
	}
	duration, err := upnpav.ParseDuration(rsp.Duration)
	if err != nil {
		return rsp.URI, didl, 0, 0, fmt.Errorf("could not unmarshal duration: %w", err)
	}
	progress, err := upnpav.ParseDuration(rsp.RelativeTime)
	if err != nil {
		return rsp.URI, didl, duration, 0, fmt.Errorf("could not unmarshal duration: %w", err)
	}

	return rsp.URI, didl, duration, progress, nil
}
func (c *client) TransportInfo(ctx context.Context) (State, error) {
	req := getTransportInfoRequest{}
	rsp := getTransportInfoResponse{}
	if err := c.call(ctx, getTransportInfo, req, &rsp); err != nil {
		return State(""), err
	}
	return rsp.TransportState, nil
}

func (c *client) SetCurrentURI(ctx context.Context, uri string, metadata *upnpav.DIDL) error {
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
func (c *client) SetNextURI(ctx context.Context, uri string, metadata *upnpav.DIDL) error {
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
func marshalAndMaybeCreateDIDL(uri string, metadata *upnpav.DIDL) ([]byte, error) {
	if metadata == nil {
		didl, err := upnpav.DIDLForURI(uri)
		if err != nil {
			return nil, fmt.Errorf("could not create DIDL for URI %v: %w", uri, err)
		}
		metadata = didl
	}
	//metadata.Items[0].Resources[0].ProtocolInfo.AdditionalInfo = "DLNA.ORG_PN=AVC_MP4_HP_HD_AAC;DLNA.ORG_OP=01;DLNA.ORG_CI=0;DLNA.ORG_FLAGS=01700000000000000000000000000000"
	metadataBytes, err := xml.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("could not marshal DIDL to string: %w", err)
	}
	return metadataBytes, nil
}
