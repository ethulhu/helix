package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
)

type (
	Queue struct {
		mu sync.Mutex

		state avtransport.State

		// queue TrackSequence
		queue *TrackList

		transport avtransport.Client
		udn       string
		name      string
		sinks     []*upnpav.ProtocolInfo
	}

	transportState struct {
		state avtransport.State
		uri   string
		didl  *upnpav.DIDL
	}
)

func NewQueue() *Queue {
	q := &Queue{
		state: avtransport.StateStopped,
		queue: &TrackList{},
	}

	go func() {
		ctx := context.Background()

		var prevTS *transportState
		ts := &transportState{
			state: avtransport.StateStopped,
		}

		for _ = range time.Tick(1 * time.Second) {
			if q.transport == nil {
				continue
			}

			prevTS = ts

			var err error
			ts, err = q.transportState(ctx)
			if err != nil {
				log.Printf("could not get transport state: %v", err)
				continue
			}

			if ts.state == avtransport.StateTransitioning {
				continue
			}
			if q.state == avtransport.StateStopped && ts.state == q.state {
				continue
			}
			if q.state == avtransport.StatePaused && ts.state == q.state {
				continue
			}

			if q.state == avtransport.StateStopped {
				if err := q.transport.Stop(ctx); err != nil {
					log.Printf("could not stop transport: %v", err)
				}
				continue
			}
			if q.state == avtransport.StatePaused {
				if err := q.transport.Pause(ctx); err != nil {
					log.Printf("could not pause transport: %v", err)
				}
				continue
			}

			// q.state can only be avtransport.StatePlaying from here.

			current, ok := q.queue.Current()
			if !ok {
				continue
			}
			currentURI, ok := current.URIForProtocolInfos(q.sinks)
			if !ok {
				// If the transport can't play it, skip again.
				q.queue.Skip()
				continue
			}

			if ts.uri == currentURI && ts.state == avtransport.StatePlaying {
				// Everything is OK.
				continue
			}
			if ts.uri == currentURI && ts.state == avtransport.StatePaused {
				// Unpause.
				if err := q.transport.Play(ctx); err != nil {
					log.Printf("could not play transport: %v", err)
				}
				continue
			}

			if prevTS.state == avtransport.StatePlaying && prevTS.uri == currentURI {
				// We've fallen behind, so skip ourselves.
				q.queue.Skip()
				var ok bool
				current, ok = q.queue.Current()
				if !ok {
					continue
				}
				currentURI, ok = current.URIForProtocolInfos(q.sinks)
				if !ok {
					// If the transport can't play it, skip again.
					q.queue.Skip()
					continue
				}
			}

			didl := &upnpav.DIDL{Items: []upnpav.Item{current}}
			if ts.state != avtransport.StateStopped {
				_ = q.transport.Stop(ctx)
			}
			if err := q.transport.SetCurrentURI(ctx, currentURI, didl); err != nil {
				log.Printf("could not set transport URI: %v", err)
			}
			if err := q.transport.Play(ctx); err != nil {
				log.Printf("could not play transport: %v", err)
			}
		}
	}()

	return q
}

func (q *Queue) transportState(ctx context.Context) (*transportState, error) {
	if q.transport == nil {
		return nil, nil
	}
	state, _, err := q.transport.TransportInfo(ctx)
	if err != nil {
		return nil, err
	}
	t := &transportState{state: state}
	if state != avtransport.StateStopped {
		uri, didl, _, _, err := q.transport.PositionInfo(ctx)
		if err != nil {
			return t, nil
		}
		t.uri = uri
		t.didl = didl
	}
	return t, nil
}

func (q *Queue) Play()  { q.state = avtransport.StatePlaying }
func (q *Queue) Pause() { q.state = avtransport.StatePaused }
func (q *Queue) Stop()  { q.state = avtransport.StateStopped }

func (q *Queue) SetTransport(device *ssdp.Device) error {
	connMgr, ok := device.Client(connectionmanager.Version1)
	if !ok {
		return errors.New("device does not expose ConnectionManager")
	}
	transport, ok := device.Client(avtransport.Version1)
	if !ok {
		return errors.New("device does not expose AVTransport")
	}

	ctx := context.Background()
	_, sinks, err := connectionmanager.NewClient(connMgr).ProtocolInfo(ctx)
	if err != nil {
		return fmt.Errorf("could not list device sink protocols: %w", err)
	}
	if len(sinks) == 0 {
		return errors.New("device has no valid sink protocols")
	}

	q.name = device.Name
	q.sinks = sinks
	q.transport = avtransport.NewClient(transport)
	q.udn = device.UDN
	return nil
}
func (q *Queue) AddLast(item upnpav.Item) {
	q.queue.AddLast(item)
}

func (q *Queue) Clear() {
	q.queue.Clear()
}

func (q *Queue) State() avtransport.State {
	return q.state
}
func (q *Queue) Queue() []upnpav.Item {
	return q.queue.Items()
}
