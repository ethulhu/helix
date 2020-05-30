// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package avtransport

import (
	"context"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

type (
	// Interface is the UPnP AVTransport:1 interface.
	// Not all methods exist on all Renderers.
	// For example, Next is missing on gmediarender.
	Interface interface {
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
		SetCurrentURI(ctx context.Context, uri string, metadata *upnpav.DIDLLite) error
		// SetNextURI sets the URI of the next track.
		// If metadata is nil, it will create a minimal metadata.
		SetNextURI(ctx context.Context, uri string, metadata *upnpav.DIDLLite) error

		// MediaInfo returns the current URI and metadata.
		MediaInfo(context.Context) (string, *upnpav.DIDLLite, string, *upnpav.DIDLLite, error)

		// PositionInfo returns the current URI, metadata, total time, and elapsed time.
		PositionInfo(context.Context) (string, *upnpav.DIDLLite, time.Duration, time.Duration, error)

		// TransportInfo returns the current playback state and error status.
		TransportInfo(context.Context) (State, Status, error)
	}

	SeekMode string

	// State is a playback state.
	// Vendor defined states can exist.
	State string

	// Status is whether or not an error occurred.
	// Vendor defined statuses can exist.
	Status string
)

const (
	Version1 = ssdp.URN("urn:schemas-upnp-org:service:AVTransport:1")
	Version2 = ssdp.URN("urn:schemas-upnp-org:service:AVTransport:2")
)

const (
	// The spec requires AVTransports to support StatePlaying & StateStopped.

	StatePlaying = State("PLAYING")
	StateStopped = State("STOPPED")

	// The spec considers the rest as optional.

	StateNoMediaPresent  = State("NO_MEDIA_PRESENT")
	StatePaused          = State("PAUSED_PLAYBACK")
	StatePausedRecording = State("PAUSED_RECORDING")
	StateRecording       = State("RECORDING")
	StateTransitioning   = State("TRANSITIONING")
)

const (
	SeekRelativeTime = SeekMode("REL_TIME")
	SeekAbsoluteTime = SeekMode("ABS_TIME")
)

const (
	StatusOK    = Status("OK")
	StatusError = Status("ERROR_OCCURRED")
)
