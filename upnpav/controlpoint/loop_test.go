package controlpoint

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
)

func TestLoop(t *testing.T) {
	tests := []struct {
		comment string

		prevObservedState transportState
		currObservedState transportState
		desiredState      avtransport.State
		queueItems        []upnpav.Item
		transportChanged  bool

		wantDesiredState     avtransport.State
		wantTransportActions []string
		wantManagerActions   []string
	}{
		{
			comment: "transitioning is ignored",

			currObservedState: transportState{
				state: avtransport.StateTransitioning,
			},
		},
		{
			comment: "stopped to stopped",

			prevObservedState: transportState{},
			currObservedState: transportState{
				state: avtransport.StateStopped,
			},
			desiredState: avtransport.StateStopped,
			queueItems:   []upnpav.Item{},

			wantDesiredState:     avtransport.StateStopped,
			wantTransportActions: nil,
			wantManagerActions:   nil,
		},
		{
			comment: "playing to stopped",

			prevObservedState: transportState{},
			currObservedState: transportState{
				state: avtransport.StatePlaying,
			},
			desiredState: avtransport.StateStopped,
			queueItems:   []upnpav.Item{},

			wantDesiredState:     avtransport.StateStopped,
			wantTransportActions: []string{"stop"},
			wantManagerActions:   nil,
		},
		{
			comment: "paused to paused",

			prevObservedState: transportState{},
			currObservedState: transportState{
				state: avtransport.StatePaused,
			},
			desiredState: avtransport.StatePaused,
			queueItems:   []upnpav.Item{},

			wantDesiredState:     avtransport.StatePaused,
			wantTransportActions: nil,
			wantManagerActions:   nil,
		},
		{
			comment: "playing to paused",

			prevObservedState: transportState{},
			currObservedState: transportState{
				state: avtransport.StatePlaying,
			},
			desiredState: avtransport.StatePaused,
			queueItems:   []upnpav.Item{},

			wantDesiredState:     avtransport.StatePaused,
			wantTransportActions: []string{"pause"},
			wantManagerActions:   nil,
		},
		{
			comment: "external control restarted playback",

			prevObservedState: transportState{
				state: avtransport.StatePaused,
			},
			currObservedState: transportState{
				state: avtransport.StatePlaying,
			},
			desiredState: avtransport.StatePaused,
			queueItems:   []upnpav.Item{},

			wantDesiredState:     avtransport.StatePlaying,
			wantTransportActions: nil,
			wantManagerActions:   nil,
		},
		{
			// TODO: this should take seek into account.
			comment: "playing to playing same URI",

			currObservedState: transportState{
				state: avtransport.StatePlaying,
				uri:   "http://mew/purr.mp3",
			},
			desiredState: avtransport.StatePlaying,
			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					{
						URI: "http://mew/purr.mp3",
					},
				}},
			},

			wantDesiredState:     avtransport.StatePlaying,
			wantTransportActions: nil,
			wantManagerActions:   nil,
		},
		{
			comment: "external control paused playback",

			prevObservedState: transportState{
				state: avtransport.StatePlaying,
			},
			currObservedState: transportState{
				state: avtransport.StatePaused,
			},
			desiredState: avtransport.StatePlaying,
			queueItems:   []upnpav.Item{},

			wantDesiredState:     avtransport.StatePaused,
			wantTransportActions: nil,
			wantManagerActions:   nil,
		},
		{
			comment: "playing to stopped to playing",

			prevObservedState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 5 * time.Second,
			},
			currObservedState: transportState{
				state: avtransport.StateStopped,
			},
			desiredState: avtransport.StatePlaying,
			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},

			wantDesiredState:     avtransport.StatePlaying,
			wantTransportActions: []string{"setCurrentURI http://mew/purr2.mp3", "play", "seek 0"},
			wantManagerActions:   []string{"protocolInfo"},
		},
		{
			comment: "playing to playing same URI on new transport",

			prevObservedState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 5 * time.Second,
			},
			currObservedState: transportState{
				state: avtransport.StateStopped,
			},
			desiredState: avtransport.StatePlaying,
			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},
			transportChanged: true,

			wantDesiredState:     avtransport.StatePlaying,
			wantTransportActions: []string{"setCurrentURI http://mew/purr1.mp3", "play", "seek 5"},
			wantManagerActions:   []string{"protocolInfo"},
		},
		{
			comment: "playing to skipping to playing on new transport",

			prevObservedState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 5 * time.Second,
			},
			currObservedState: transportState{
				state: avtransport.StateStopped,
			},
			desiredState: avtransport.StatePlaying,
			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/flac"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},
			transportChanged: true,

			wantDesiredState:     avtransport.StatePlaying,
			wantTransportActions: []string{"setCurrentURI http://mew/purr2.mp3", "play", "seek 0"},
			wantManagerActions:   []string{"protocolInfo"},
		},
		{
			comment: "playing to stopped to skipping to playing",

			prevObservedState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 5 * time.Second,
			},
			currObservedState: transportState{
				state: avtransport.StateStopped,
			},
			desiredState: avtransport.StatePlaying,
			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/flac"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr3.mp3", "audio/mpeg"),
				}},
			},

			wantDesiredState:     avtransport.StatePlaying,
			wantTransportActions: []string{"setCurrentURI http://mew/purr3.mp3", "play", "seek 0"},
			wantManagerActions:   []string{"protocolInfo"},
		},
		{
			comment: "playing to stopped to skipping to playing",

			prevObservedState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 5 * time.Second,
			},
			currObservedState: transportState{
				state: avtransport.StateStopped,
			},
			desiredState: avtransport.StatePlaying,
			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr1.mp3", "audio/mpeg"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/flac"),
				}},
				{Resources: []upnpav.Resource{
					resource("http://mew/purr3.mp3", "audio/mpeg"),
				}},
			},

			wantDesiredState:     avtransport.StatePlaying,
			wantTransportActions: []string{"setCurrentURI http://mew/purr3.mp3", "play", "seek 0"},
			wantManagerActions:   []string{"protocolInfo"},
		},
		{
			comment: "playing to playing a different track",

			prevObservedState: transportState{},
			currObservedState: transportState{
				state:   avtransport.StatePlaying,
				uri:     "http://mew/purr1.mp3",
				elapsed: 5 * time.Second,
			},
			desiredState: avtransport.StatePlaying,
			queueItems: []upnpav.Item{
				{Resources: []upnpav.Resource{
					resource("http://mew/purr2.mp3", "audio/mpeg"),
				}},
			},

			wantDesiredState:     avtransport.StatePlaying,
			wantTransportActions: []string{"stop", "setCurrentURI http://mew/purr2.mp3", "play", "seek 0"},
			wantManagerActions:   []string{"protocolInfo"},
		},
	}

	for i, tt := range tests {
		transport := &fakeAVTransport{}
		manager := &fakeConnectionManager{}

		queue := &TrackList{items: tt.queueItems}

		gotDesiredState, err := tick(nil, tt.prevObservedState, tt.currObservedState, transport, manager, tt.desiredState, queue, tt.transportChanged)
		if err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}

		if gotDesiredState != tt.wantDesiredState {
			t.Errorf("[%d]: got desired state %v, wanted %q", i, gotDesiredState, tt.wantDesiredState)
		}
		if !reflect.DeepEqual(transport.actions, tt.wantTransportActions) {
			t.Errorf("[%d]: got AVTransport actions %q, wanted %q", i, transport.actions, tt.wantTransportActions)
		}
		if !reflect.DeepEqual(manager.actions, tt.wantManagerActions) {
			t.Errorf("[%d]: got ConnectionManager actions %q, wanted %q", i, manager.actions, tt.wantManagerActions)
		}
	}
}

type fakeAVTransport struct {
	avtransport.Client
	actions []string
}

func (c *fakeAVTransport) Play(_ context.Context) error {
	c.actions = append(c.actions, "play")
	return nil
}
func (c *fakeAVTransport) Pause(_ context.Context) error {
	c.actions = append(c.actions, "pause")
	return nil
}
func (c *fakeAVTransport) Stop(_ context.Context) error {
	c.actions = append(c.actions, "stop")
	return nil
}
func (c *fakeAVTransport) SetCurrentURI(_ context.Context, uri string, _ *upnpav.DIDL) error {
	c.actions = append(c.actions, fmt.Sprintf("setCurrentURI %v", uri))
	return nil
}
func (c *fakeAVTransport) Seek(_ context.Context, elapsed time.Duration) error {
	c.actions = append(c.actions, fmt.Sprintf("seek %v", elapsed.Seconds()))
	return nil
}

type fakeConnectionManager struct {
	connectionmanager.Client
	actions []string
}

func (c *fakeConnectionManager) ProtocolInfo(_ context.Context) ([]*upnpav.ProtocolInfo, []*upnpav.ProtocolInfo, error) {
	c.actions = append(c.actions, "protocolInfo")
	sinks := []*upnpav.ProtocolInfo{
		{
			Protocol:       upnpav.ProtocolHTTP,
			Network:        "*",
			ContentFormat:  "audio/mpeg",
			AdditionalInfo: "*",
		},
	}
	return nil, sinks, nil
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
