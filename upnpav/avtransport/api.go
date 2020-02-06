package avtransport

import (
	"context"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

type (
	// Client is a UPnP AVTransport1 client.
	Client interface {
		// Name is the "friendly name" of the UPnP Device.
		Name() string

		// Play plays the current track.
		Play(context.Context) error
		// Pause pauses the current track.
		Pause(context.Context) error
		// Next skips to the next track.
		Next(context.Context) error
		// Previous skips back to the previous track.
		Previous(context.Context) error
		// Stop stops playback altogether.
		Stop(context.Context) error
		// Seek seeks to a given time.
		Seek(context.Context, time.Duration) error

		// SetCurrentURI sets the URI of the current track.
		// If metadata is nil, it will create a minimal metadata.
		SetCurrentURI(ctx context.Context, uri string, metadata *upnpav.DIDL) error
		// SetNextURI sets the URI of the next track.
		// If metadata is nil, it will create a minimal metadata.
		SetNextURI(ctx context.Context, uri string, metadata *upnpav.DIDL) error

		MediaInfo(context.Context) (string, *upnpav.DIDL, error)
		PositionInfo(context.Context) (string, *upnpav.DIDL, time.Duration, time.Duration, error)
		TransportInfo(context.Context) (State, error)
	}

	State    string
	SeekMode string
)

const (
	Version1 = ssdp.URN("urn:schemas-upnp-org:service:AVTransport:1")
	Version2 = ssdp.URN("urn:schemas-upnp-org:service:AVTransport:2")
)

const (
	StatePaused  = State("PAUSED_PLAYBACK")
	StatePlaying = State("PLAYING")
	StateStopped = State("STOPPED")
)

const (
	SeekRelativeTime = SeekMode("REL_TIME")
	SeekAbsoluteTime = SeekMode("ABS_TIME")
)
