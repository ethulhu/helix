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

		queue   []upnpav.Item
		current *upnpav.Item
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
	}

	go func() {
		ctx := context.Background()
		for _ = range time.Tick(1 * time.Second) {
			if q.transport == nil {
				continue
			}

			state, err := q.transport.TransportInfo(ctx)
			if err != nil {
				log.Printf("could not get transport state: %v", err)
				continue
			}

			if q.state == avtransport.StateStopped && state == q.state {
				continue
			}
			if q.state == avtransport.StatePaused && state == q.state {
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

			if len(q.queue) == 0 {
				continue
			}

			if state == avtransport.StateStopped {
				if q.current != nil {
					// Skip to the next track.
					q.queue = q.queue[1:]
					if len(q.queue) == 0 {
						// We ran out of tracks.
						q.current = nil
						continue
					}
				}
				q.current = &q.queue[0]
				currentURI := q.current.Resources[0].URI
				currentDIDL := &upnpav.DIDL{Items: []upnpav.Item{*q.current}}

				if err := q.transport.SetCurrentURI(ctx, currentURI, currentDIDL); err != nil {
					log.Printf("could not set transport URI: %v", err)
					continue
				}
				if err := q.transport.Play(ctx); err != nil {
					log.Printf("could not play transport: %v", err)
					continue
				}
				continue
			}

			uri, _, _, _, err := q.transport.PositionInfo(ctx)
			if err != nil {
				log.Printf("could not get transport position: %v", err)
				continue
			}
			if uri == q.current.Resources[0].URI && state == avtransport.StatePlaying {
				// Everything is OK.
				continue
			}
			if uri == q.current.Resources[0].URI && state == avtransport.StatePaused {
				// Unpause.
				if err := q.transport.Play(ctx); err != nil {
					log.Printf("could not play transport: %v", err)
				}
				continue
			}

			currentURI := q.current.Resources[0].URI
			currentDIDL := &upnpav.DIDL{Items: []upnpav.Item{*q.current}}
			// transport state is either Playing or Paused,
			// either way we stop it, set the new URI, then start it again.
			_ = q.transport.Stop(ctx)
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

func (q *Queue) Play()  { q.state = avtransport.StatePlaying }
func (q *Queue) Pause() { q.state = avtransport.StatePaused }
func (q *Queue) Stop()  { q.state = avtransport.StateStopped }

func (q *Queue) SetTransport(transport avtransport.Client) {
	q.transport = transport
}
func (q *Queue) AddLast(item upnpav.Item) {
	q.queue = append(q.queue, item)
}

func (q *Queue) Clear() {
	q.queue = nil
	q.current = nil
}

func (q *Queue) State() avtransport.State {
	return q.state
}
func (q *Queue) Queue() []upnpav.Item {
	return q.queue
}
