// Package controlpoint is a UPnP AV "Control Point", for mediating ContentDirectories and AVTransports.
package controlpoint

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
)

type (
	ControlLoop struct {
		state     avtransport.State
		transport *ssdp.Device
		queue     Queue
	}
	transportState struct {
		transport *ssdp.Device
		state     avtransport.State
		uri       string
		didl      *upnpav.DIDL
		elapsed   time.Duration
	}
)

func NewControlLoop() *ControlLoop {
	cl := &ControlLoop{
		state: avtransport.StateStopped,
	}

	ctx := context.Background()
	go cl.loop(ctx)
	return cl
}

func (cl *ControlLoop) State() avtransport.State { return cl.state }

func (cl *ControlLoop) Play()  { cl.state = avtransport.StatePlaying }
func (cl *ControlLoop) Pause() { cl.state = avtransport.StatePaused }
func (cl *ControlLoop) Stop()  { cl.state = avtransport.StateStopped }

func (cl *ControlLoop) Queue() Queue {
	return cl.queue
}
func (cl *ControlLoop) SetQueue(queue Queue) {
	cl.queue = queue
}

func (cl *ControlLoop) Transport() *ssdp.Device {
	return cl.transport
}
func (cl *ControlLoop) SetTransport(device *ssdp.Device) error {
	if device == nil {
		cl.transport = nil
		return nil
	}

	if _, ok := device.SOAPClient(avtransport.Version1); !ok {
		return errors.New("device does not support AVTransport")
	}
	if _, ok := device.SOAPClient(connectionmanager.Version1); !ok {
		return errors.New("device does not support ConnectionManager")
	}

	cl.transport = device
	return nil
}

// clients will panic if cl.transport is invalid because SetTransport should make that impossible.
func clients(device *ssdp.Device) (avtransport.Client, connectionmanager.Client) {
	transportClient, ok := device.SOAPClient(avtransport.Version1)
	if !ok {
		panic(fmt.Sprintf("transport does not support AVTransport"))
	}
	managerClient, ok := device.SOAPClient(connectionmanager.Version1)
	if !ok {
		panic(fmt.Sprintf("transport does not support ConnectionManager"))
	}
	return avtransport.NewClient(transportClient), connectionmanager.NewClient(managerClient)
}

func (cl *ControlLoop) loop(ctx context.Context) {
	var prev *transportState
	curr, _ := cl.transportState(ctx)

	for _ = range time.Tick(1 * time.Second) {
		prev = curr

		// If the transport just changed, stop the old transport.
		if cl.transport != prev.transport && prev.transport != nil {
			transport, _ := clients(prev.transport)
			_ = transport.Stop(ctx)
		}

		if cl.transport == nil {
			prev.transport = nil
			continue
		}

		// TODO: instead have a NilQueue that always returns end?
		if cl.queue == nil {
			continue
		}

		transport, manager := clients(cl.transport)

		var err error
		curr, err = cl.transportState(ctx)
		if err != nil {
			log.Printf("could not get transport state: %v", err)
			continue
		}

		if curr.state == avtransport.StateTransitioning {
			continue
		}
		if cl.state == avtransport.StateStopped && curr.state == cl.state {
			continue
		}
		if cl.state == avtransport.StatePaused && curr.state == cl.state {
			continue
		}

		if cl.state == avtransport.StateStopped {
			if err := transport.Stop(ctx); err != nil {
				log.Printf("could not stop transport: %v", err)
			}
			continue
		}
		if cl.state == avtransport.StatePaused {
			// if previously it was Paused but now is Playing, and all other fields are correct, set ourseves to play.
			// TODO: check URI & elapsed time?
			if prev.state == avtransport.StatePaused && curr.state == avtransport.StatePlaying {
				cl.state = avtransport.StatePlaying
			} else {
				if err := transport.Pause(ctx); err != nil {
					log.Printf("could not pause transport: %v", err)
				}
			}
			continue
		}
		if cl.state == avtransport.StatePlaying && prev.state == avtransport.StatePlaying && curr.state == avtransport.StatePaused {
			// TODO: check URI & elapsed time?
			cl.state = avtransport.StatePaused
			continue
		}

		// cl.state can only be avtransport.StatePlaying from here.

		currentItem, ok := cl.queue.Current()
		if !ok {
			// We've run out of tracks.
			continue
		}

		if currentItem.HasURI(curr.uri) && curr.state == avtransport.StatePlaying {
			// Everything is OK.
			continue
		}
		if currentItem.HasURI(curr.uri) && curr.state == avtransport.StatePaused {
			// Unpause.
			if err := transport.Play(ctx); err != nil {
				log.Printf("could not play transport: %v", err)
			}
			continue
		}

		_, sinks, err := manager.ProtocolInfo(ctx)
		if err != nil {
			log.Printf("could not get sink protocols for device: %v", err)
			continue
		}
		if len(sinks) == 0 {
			log.Print("got 0 sink protocols for device, expected at least 1")
			continue
		}

		newTrack := false
		if currentItem.HasURI(prev.uri) && prev.state == avtransport.StatePlaying {
			// We've fallen behind, so skip ourselves.
			log.Print("skipping track")
			cl.queue.Skip()
			newTrack = true

			var ok bool
			currentItem, ok = cl.queue.Current()
			if !ok {
				// We ran out of tracks.
				continue
			}
		}

		didl := &upnpav.DIDL{Items: []upnpav.Item{currentItem}}
		// TODO: replace with a loop.
		currentURI, ok := currentItem.URIForProtocolInfos(sinks)
		if !ok {
			// The transport can't play it, so skip.
			log.Print("skipping track for invalid protocol")
			cl.queue.Skip()
			continue
		}

		if curr.state != avtransport.StateStopped {
			_ = transport.Stop(ctx)
		}
		if err := transport.SetCurrentURI(ctx, currentURI, didl); err != nil {
			log.Printf("could not set transport URI: %v", err)
		}
		if err := transport.Play(ctx); err != nil {
			log.Printf("could not play transport: %v", err)
		}
		if !newTrack {
			log.Printf("seeking to %v", prev.elapsed)
			if err := transport.Seek(ctx, prev.elapsed); err != nil {
				log.Printf("could not seek: %v", err)
			}
		}
	}
}
func (cl *ControlLoop) transportState(ctx context.Context) (*transportState, error) {
	t := &transportState{
		transport: cl.transport,
	}

	if cl.transport == nil {
		return t, nil
	}
	transport, _ := clients(cl.transport)

	state, _, err := transport.TransportInfo(ctx)
	if err != nil {
		return nil, err
	}
	t.state = state

	if state != avtransport.StateStopped {
		uri, didl, _, elapsed, err := transport.PositionInfo(ctx)
		if err != nil {
			return t, nil
		}
		t.uri = uri
		t.didl = didl
		t.elapsed = elapsed
	}
	return t, nil
}
