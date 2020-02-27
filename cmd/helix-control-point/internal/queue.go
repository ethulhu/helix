package internal

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
)

type (
	Queue struct {
		mu sync.Mutex

		state     avtransport.State
		transport avtransport.Client

		// queue TrackSequence
		queue *TrackList
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
			currentURI := current.Resources[0].URI
			currentDIDL := &upnpav.DIDL{Items: []upnpav.Item{current}}

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
				current, ok := q.queue.Current()
				if !ok {
					continue
				}
				currentURI = current.Resources[0].URI
				currentDIDL = &upnpav.DIDL{Items: []upnpav.Item{current}}
			}

			if ts.state != avtransport.StateStopped {
				_ = q.transport.Stop(ctx)
			}
			if err := q.transport.SetCurrentURI(ctx, currentURI, currentDIDL); err != nil {
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

func (q *Queue) SetTransport(transport avtransport.Client) {
	q.transport = transport
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
