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
	Loop struct {
		state     avtransport.State
		transport *ssdp.Device
		queue     Queue
	}
	transportState struct {
		state   avtransport.State
		uri     string
		elapsed time.Duration
	}
)

func NewLoop() *Loop {
	loop := &Loop{
		state: avtransport.StateStopped,
	}

	ctx := context.Background()
	go func() {
		// We're using UDNs instead of pointer equality for the case.
		prevDevice := loop.transport
		// prevTransport, _ := clients(prevDevice)
		// prevObservedState, err := newTransportState(ctx, prevTransport)
		prevObservedState, err := newTransportState(ctx, nil)
		if err != nil {
			log.Printf("could not get initial transport state: %v", err)
		}

		for _ = range time.Tick(1 * time.Second) {
			prevUDN := udnOrDefault(prevDevice, "")
			currUDN := udnOrDefault(loop.transport, "")

			transportChanged := currUDN != prevUDN

			if transportChanged && prevDevice != nil {
				log.Print("transport changed, stopping old transport")
				go func() {
					prevTransport, _ := clients(prevDevice)
					if err := prevTransport.Stop(ctx); err != nil {
						log.Printf("could not stop old transport: %v", err)
						return
					}
					log.Print("stopped old transport")
				}()
			}

			if loop.transport != nil {
				currTransport, currManager := clients(loop.transport)
				currObservedState, err := newTransportState(ctx, currTransport)

				newDesiredState, err := tick(ctx, prevObservedState, currObservedState, currTransport, currManager, loop.state, loop.queue, transportChanged)
				if err != nil {
					log.Print(err.Error())
				} else {
					loop.state = newDesiredState
				}
				prevObservedState = currObservedState
			}

			prevDevice = loop.transport
		}
	}()
	return loop
}

func (loop *Loop) State() avtransport.State { return loop.state }

func (loop *Loop) Play()  { loop.state = avtransport.StatePlaying }
func (loop *Loop) Pause() { loop.state = avtransport.StatePaused }
func (loop *Loop) Stop()  { loop.state = avtransport.StateStopped }

func (loop *Loop) Queue() Queue {
	return loop.queue
}
func (loop *Loop) SetQueue(queue Queue) {
	loop.queue = queue
}

func (loop *Loop) Transport() *ssdp.Device {
	return loop.transport
}
func (loop *Loop) SetTransport(device *ssdp.Device) error {
	if device == nil {
		loop.transport = nil
		return nil
	}

	if _, ok := device.SOAPClient(avtransport.Version1); !ok {
		return errors.New("device does not support AVTransport")
	}
	if _, ok := device.SOAPClient(connectionmanager.Version1); !ok {
		return errors.New("device does not support ConnectionManager")
	}

	loop.transport = device
	return nil
}

// clients will panic if device is invalid because SetTransport should make that impossible.
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

// tick is a 7-argument monstrosity to make it clear what it consumes.
func tick(ctx context.Context,
	prev transportState,
	curr transportState,
	transport avtransport.Client,
	manager connectionmanager.Client,
	desiredState avtransport.State,
	queue Queue,
	transportChanged bool) (avtransport.State, error) {

	if transport == nil {
		log.Print("no current transport, doing nothing")
		return desiredState, nil
	}

	if queue == nil {
		log.Print("no current queue, doing nothing")
		return desiredState, nil
	}

	if curr.state == avtransport.StateTransitioning {
		log.Printf("transport in state %v, doing nothing", avtransport.StateTransitioning)
		return desiredState, nil
	}

	switch desiredState {

	case avtransport.StateStopped:
		switch curr.state {
		case avtransport.StateStopped:
			log.Printf("transport in desired state %v, doing nothing", desiredState)
		default:
			log.Print("stopping transport")
			if err := transport.Stop(ctx); err != nil {
				return desiredState, fmt.Errorf("could not stop transport: %w", err)
			}
		}
		return desiredState, nil

	case avtransport.StatePaused:
		switch curr.state {
		case avtransport.StatePaused:
			log.Printf("transport in desired state %v, doing nothing", desiredState)
		default:
			// TODO: also check URI & elapsed time?
			if prev.state == avtransport.StatePaused && curr.state == avtransport.StatePlaying {
				log.Print("transport was previously paused, but was starting playing externally, doing nothing")
				return avtransport.StatePlaying, nil
			}

			log.Print("pausing transport")
			if err := transport.Pause(ctx); err != nil {
				return desiredState, fmt.Errorf("could not pause transport: %w", err)
			}
		}
		return desiredState, nil

	case avtransport.StatePlaying:
		// TODO: generalize to "everything else matches, and prev.state matched, but the curr.state differs"?
		if prev.state == avtransport.StatePlaying && curr.state == avtransport.StatePaused {
			log.Print("transport was previously playing, but was paused externally, doing nothing")
			return avtransport.StatePaused, nil
		}

		currentItem, ok := queue.Current()
		if !ok {
			log.Print("reached end of queue, doing nothing")
			return avtransport.StateStopped, nil
		}

		// TODO: maybe check seek?
		if currentItem.HasURI(curr.uri) {
			switch curr.state {
			case avtransport.StatePlaying:
				log.Print("current transport URI & state match, doing nothing")
				return desiredState, nil
			default:
				log.Printf("current transport URI matches but state is %v, starting playback", curr.state)
				if err := transport.Play(ctx); err != nil {
					return desiredState, fmt.Errorf("could not play transport: %w", err)
				}
				return desiredState, nil
			}
		}

		if !transportChanged && currentItem.HasURI(prev.uri) { // TODO: && prev.state == avtransport.StatePlaying ?
			log.Print("controller has fallen behind, skipping track")
			currentItem, ok = queue.Skip()
		}

		seek := prev.elapsed
		if !transportChanged && !currentItem.HasURI(curr.uri) {
			seek = 0
		}

		log.Print("getting ProtocolInfos for transport")
		_, sinks, err := manager.ProtocolInfo(ctx)
		if err != nil {
			return desiredState, fmt.Errorf("could not get sink protocols for device: %w", err)
		}
		if len(sinks) == 0 {
			return desiredState, errors.New("got 0 sink protocols for device, expected at least 1")
		}

		currentURI, ok := currentItem.URIForProtocolInfos(sinks)
		if !ok {
			for {
				currentItem, ok = queue.Skip()
				if !ok {
					// We ran out of tracks.
					break
				}
				seek = 0

				currentURI, ok = currentItem.URIForProtocolInfos(sinks)
				if ok {
					// We found a playable track.
					break
				}
			}
			if !ok {
				log.Print("reached end of queue, doing nothing")
				return avtransport.StateStopped, nil
			}
		}

		// TODO: make this less abrupt, gapless playback, etc.
		if curr.state != avtransport.StateStopped {
			log.Print("temporarily stopping transport")
			_ = transport.Stop(ctx)
		}

		log.Print("setting current transport URI")
		metadata := &upnpav.DIDL{Items: []upnpav.Item{currentItem}}
		if err := transport.SetCurrentURI(ctx, currentURI, metadata); err != nil {
			return desiredState, fmt.Errorf("could not set transport URI: %w", err)
		}

		log.Print("starting transport playing")
		if err := transport.Play(ctx); err != nil {
			return desiredState, fmt.Errorf("could not play: %w", err)
		}

		if transportChanged && seek != 0 {
			log.Printf("seeking transport to %v", seek)
			if err := transport.Seek(ctx, seek); err != nil {
				return desiredState, fmt.Errorf("could not seek: %w", err)
			}
		}
		return desiredState, nil

	default:
		panic(fmt.Sprintf("can only have desired states %v, got %q", []avtransport.State{avtransport.StatePlaying, avtransport.StatePaused, avtransport.StateStopped}, desiredState))
	}
}

func newTransportState(ctx context.Context, transport avtransport.Client) (transportState, error) {
	t := transportState{}

	if transport == nil {
		return t, nil
	}

	state, _, err := transport.TransportInfo(ctx)
	if err != nil {
		return t, err
	}
	t.state = state

	if state != avtransport.StateStopped {
		uri, _, _, elapsed, err := transport.PositionInfo(ctx)
		if err != nil {
			return t, nil
		}
		t.uri = uri
		t.elapsed = elapsed
	}
	return t, nil
}

func udnOrDefault(device *ssdp.Device, def string) string {
	if device == nil {
		return def
	}
	return device.UDN
}
