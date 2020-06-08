// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package controlpoint

import (
	"testing"
	"time"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
)

func TestLoop(t *testing.T) {
	tests := []struct {
		comment string

		queueItems         []upnpav.Item
		prevTransportState transportState
		currTransportState transportState
		loopState          avtransport.State
		loopElapsed        time.Duration
		transportChanged   bool

		wantLoopState   avtransport.State
		wantLoopElapsed time.Duration
		wantAction      action
	}{
		{
			comment: "transitioning is ignored",

			currTransportState: transportState{
				state: avtransport.StateTransitioning,
			},
			loopState:   avtransport.StatePlaying,
			loopElapsed: 1 * time.Minute,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 1 * time.Minute,
			wantAction:      doNothing,
		},
		{
			comment: "stopped to stopped",

			prevTransportState: transportState{},
			currTransportState: transportState{
				state: avtransport.StateStopped,
			},
			loopState: avtransport.StateStopped,

			wantLoopState:   avtransport.StateStopped,
			wantLoopElapsed: 0,
			wantAction:      doNothing,
		},
		{
			comment: "playing to stopped",

			prevTransportState: transportState{},
			currTransportState: transportState{
				state: avtransport.StatePlaying,
			},
			loopState:   avtransport.StateStopped,
			loopElapsed: 1 * time.Minute,

			wantLoopState:   avtransport.StateStopped,
			wantLoopElapsed: 0,
			wantAction:      stop,
		},
		{
			comment: "paused to paused",

			prevTransportState: transportState{
				state:   avtransport.StatePaused,
				elapsed: 1 * time.Minute,
			},
			currTransportState: transportState{
				state:   avtransport.StatePaused,
				elapsed: 1 * time.Minute,
			},
			loopState:   avtransport.StatePaused,
			loopElapsed: 1 * time.Minute,

			wantLoopState:   avtransport.StatePaused,
			wantLoopElapsed: 1 * time.Minute,
			wantAction:      doNothing,
		},
		{
			comment: "playing to paused",

			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				elapsed: 1*time.Minute - 1*time.Second,
			},
			currTransportState: transportState{
				state:   avtransport.StatePlaying,
				elapsed: 1 * time.Minute,
			},
			loopState:   avtransport.StatePaused,
			loopElapsed: 1*time.Minute - 1*time.Second,

			wantLoopState:   avtransport.StatePaused,
			wantLoopElapsed: 1 * time.Minute,
			wantAction:      pause,
		},
		{
			comment: "external control restarted playback",

			prevTransportState: transportState{
				state:   avtransport.StatePaused,
				elapsed: 1 * time.Minute,
			},
			currTransportState: transportState{
				state:   avtransport.StatePlaying,
				elapsed: 1 * time.Minute,
			},
			loopState:   avtransport.StatePaused,
			loopElapsed: 1 * time.Minute,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 1 * time.Minute,
			wantAction:      doNothing,
		},
		{
			// TODO: this should take seek into account.
			comment: "playing to playing same URI",

			queueItems: []upnpav.Item{{
				Resources: []upnpav.Resource{
					resource("http://mew/purr.mp3", "audio/mpeg"),
				},
			}},
			currTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr.mp3",
				elapsed: 1 * time.Minute,
			},
			loopState:   avtransport.StatePlaying,
			loopElapsed: 1 * time.Minute,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 1 * time.Minute,
			wantAction:      doNothing,
		},
		{
			comment: "external control paused playback",

			queueItems: []upnpav.Item{},
			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				elapsed: 1 * time.Minute,
			},
			currTransportState: transportState{
				state:   avtransport.StatePaused,
				elapsed: 1 * time.Minute,
			},
			loopState:   avtransport.StatePlaying,
			loopElapsed: 1 * time.Minute,

			wantLoopState:   avtransport.StatePaused,
			wantLoopElapsed: 1 * time.Minute,
			wantAction:      doNothing,
		},
		{
			comment: "playing to stopped to playing on same transport",

			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},
			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1 * time.Minute,
			},
			currTransportState: transportState{
				state: avtransport.StateStopped,
			},
			transportChanged: false,
			loopState:        avtransport.StatePlaying,
			loopElapsed:      1 * time.Minute,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 0,
			wantAction:      skipTrack,
		},
		{
			comment: "playing to playing on same transport (e.g. SetNextURI)",

			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},
			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1 * time.Minute,
			},
			currTransportState: transportState{
				state: avtransport.StatePlaying,
				uri:   "http://mew/purr2.mp3",
			},
			transportChanged: false,
			loopState:        avtransport.StatePlaying,
			loopElapsed:      1 * time.Minute,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 0,
			wantAction:      skipTrack,
		},
		{
			comment: "playing to playing same URI on new transport",
			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},
			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1 * time.Minute,
			},
			currTransportState: transportState{
				state: avtransport.StateStopped,
			},
			transportChanged: true,
			loopState:        avtransport.StatePlaying,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 1 * time.Minute,
			wantAction:      setURI,
		},
		{
			comment: "playing to skipping to playing on new transport",

			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.flac", "audio/flac"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},
			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.flac",
				elapsed: 1 * time.Minute,
			},
			currTransportState: transportState{
				state: avtransport.StateStopped,
			},
			loopState:        avtransport.StatePlaying,
			transportChanged: true,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 0,
			wantAction:      skipTrack,
		},
		{
			comment: "playing an empty queue to stopped",

			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr.mp3",
				elapsed: 1*time.Minute - 1*time.Second,
			},
			currTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr.mp3",
				elapsed: 1 * time.Minute,
			},
			loopState:   avtransport.StatePlaying,
			loopElapsed: 1*time.Minute - 1*time.Second,
			queueItems:  nil,

			wantLoopState:   avtransport.StateStopped,
			wantLoopElapsed: 0,
			wantAction:      stop,
		},
		{
			comment: "playing to playing a different track",

			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},
			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1*time.Minute - 1*time.Second,
			},
			currTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1 * time.Minute,
			},
			loopState:        avtransport.StatePlaying,
			loopElapsed:      1 * time.Minute,
			transportChanged: false,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 0,
			wantAction:      setURI,
		},
		{
			comment: "playing to seeking forward",

			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
			},
			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1*time.Minute - 1*time.Second,
			},
			currTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1 * time.Minute,
			},
			loopState:   avtransport.StatePlaying,
			loopElapsed: 2 * time.Minute,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 2 * time.Minute,
			wantAction:      seek,
		},
		{
			comment: "playing to seeking backward",

			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
			},
			prevTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1*time.Minute - 1*time.Second,
			},
			currTransportState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 1 * time.Minute,
			},
			loopState:   avtransport.StatePlaying,
			loopElapsed: 30 * time.Second,

			wantLoopState:   avtransport.StatePlaying,
			wantLoopElapsed: 30 * time.Second,
			wantAction:      seek,
		},
	}

	for i, tt := range tests {
		protocolInfos := []upnpav.ProtocolInfo{
			{
				Protocol:       upnpav.ProtocolHTTP,
				Network:        "*",
				ContentFormat:  "audio/mpeg",
				AdditionalInfo: "*",
			},
		}

		queue := NewTrackList()
		for _, item := range tt.queueItems {
			queue.Append(item)
		}

		gotLoopState, gotLoopElapsed, gotAction := tick(queue, protocolInfos, tt.prevTransportState, tt.currTransportState, tt.loopState, tt.loopElapsed, tt.transportChanged)

		if gotLoopState != tt.wantLoopState {
			t.Errorf("[%d]: got loop state %v, wanted %v", i, gotLoopState, tt.wantLoopState)
		}
		if gotLoopElapsed != tt.wantLoopElapsed {
			t.Errorf("[%d]: got loop elapsed %v, wanted %v", i, gotLoopElapsed, tt.wantLoopElapsed)
		}
		if gotAction != tt.wantAction {
			t.Errorf("[%d]: got action %v, wanted %v", i, gotAction, tt.wantAction)
		}
	}
}

func resource(uri, mime string) upnpav.Resource {
	return upnpav.Resource{
		URI: uri,
		ProtocolInfo: &upnpav.ProtocolInfo{
			Protocol:       upnpav.ProtocolHTTP,
			Network:        "*",
			ContentFormat:  mime,
			AdditionalInfo: "*",
		},
	}
}
