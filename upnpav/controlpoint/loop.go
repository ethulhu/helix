// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Package controlpoint is a UPnP AV "Control Point", for mediating ContentDirectories and AVTransports.
package controlpoint

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethulhu/helix/logger"
	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
)

type (
	Loop struct {
		state     avtransport.State
		transport *upnp.Device
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
		prevObservedState, err := newTransportState(ctx, nil)
		if err != nil {
			// We passed a nil transport, so it shouldn't be possible to get errors here.
			panic(fmt.Sprintf("could not get initial transport state: %v", err))
		}

		for _ = range time.Tick(1 * time.Second) {
			prevUDN := udnOrDefault(prevDevice, "")
			currUDN := udnOrDefault(loop.transport, "")

			transportChanged := currUDN != prevUDN

			if transportChanged && prevDevice != nil {
				go func() {
					log, ctx := logger.FromContext(ctx)
					log.AddField("transport.previous.udn", prevDevice.UDN)
					log.AddField("transport.previous.name", prevDevice.Name)

					prevTransport, _ := clients(prevDevice)
					if err := prevTransport.Stop(ctx); err != nil {
						log.WithError(err).Warning("could not stop previous transport")
						return
					}
					log.Info("stopped previous transport")
				}()
			}

			if loop.transport != nil {
				log, ctx := logger.FromContext(ctx)
				log.AddField("transport.udn", loop.transport.UDN)
				log.AddField("transport.name", loop.transport.Name)

				currTransport, currManager := clients(loop.transport)
				currObservedState, err := newTransportState(ctx, currTransport)
				if err != nil {
					log.WithError(err).Warning("could not get transport state")
					continue
				}

				log.AddField("current.state", currObservedState.state)
				if currObservedState.state == avtransport.StatePlaying || currObservedState.state == avtransport.StatePaused {
					log.AddField("current.uri", currObservedState.uri)
				}

				newDesiredState, err := tick(ctx, prevObservedState, currObservedState, currTransport, currManager, loop.state, loop.queue, transportChanged)
				if err != nil {
					log.WithError(err).Warning("error determining new loop state")
				} else {
					if loop.state != newDesiredState {
						loop.state = newDesiredState
						log.AddField("new.state", newDesiredState)
						log.Info("updated desired loop state")
					}
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

func (loop *Loop) Transport() *upnp.Device {
	return loop.transport
}
func (loop *Loop) SetTransport(device *upnp.Device) error {
	if device == nil {
		loop.transport = nil
		return nil
	}

	if _, ok := device.SOAPInterface(avtransport.Version1); !ok {
		return errors.New("device does not support AVTransport")
	}
	if _, ok := device.SOAPInterface(connectionmanager.Version1); !ok {
		return errors.New("device does not support ConnectionManager")
	}

	loop.transport = device
	return nil
}

// clients will panic if device is invalid because SetTransport should make that impossible.
func clients(device *upnp.Device) (avtransport.Interface, connectionmanager.Interface) {
	transportClient, ok := device.SOAPInterface(avtransport.Version1)
	if !ok {
		panic(fmt.Sprintf("transport does not support AVTransport"))
	}
	managerClient, ok := device.SOAPInterface(connectionmanager.Version1)
	if !ok {
		panic(fmt.Sprintf("transport does not support ConnectionManager"))
	}
	return avtransport.NewClient(transportClient), connectionmanager.NewClient(managerClient)
}

// tick is a 7-argument monstrosity to make it clear what it consumes.
func tick(ctx context.Context,
	prev transportState,
	curr transportState,
	transport avtransport.Interface,
	manager connectionmanager.Interface,
	desiredState avtransport.State,
	queue Queue,
	transportChanged bool) (avtransport.State, error) {

	log, ctx := logger.FromContext(ctx)

	if transport == nil {
		log.Info("no current transport, doing nothing")
		return desiredState, nil
	}

	if queue == nil {
		log.Info("no current queue, doing nothing")
		return desiredState, nil
	}

	if curr.state == avtransport.StateTransitioning {
		log.Info(fmt.Sprintf("transport in state %v, doing nothing", avtransport.StateTransitioning))
		return desiredState, nil
	}

	switch desiredState {

	case avtransport.StateStopped:
		switch curr.state {
		case avtransport.StateStopped:
			log.Info("transport in desired state, doing nothing")
		default:
			if err := transport.Stop(ctx); err != nil {
				log.WithError(err).Error("could not stop transport")
				return desiredState, fmt.Errorf("could not stop transport: %w", err)
			}
			log.Info("stopped transport")
		}
		return desiredState, nil

	case avtransport.StatePaused:
		switch curr.state {
		case avtransport.StatePaused:
			log.Info("transport in desired state, doing nothing")
		default:
			// TODO: also check URI & elapsed time?
			if prev.state == avtransport.StatePaused && curr.state == avtransport.StatePlaying {
				log.Info("transport was previously paused, but was started playing externally, doing nothing")
				return avtransport.StatePlaying, nil
			}

			if err := transport.Pause(ctx); err != nil {
				log.WithError(err).Error("could not pause transport")
				return desiredState, fmt.Errorf("could not pause transport: %w", err)
			}
			log.Info("paused transport")
		}
		return desiredState, nil

	case avtransport.StatePlaying:
		// TODO: generalize to "everything else matches, and prev.state matched, but the curr.state differs"?
		if prev.state == avtransport.StatePlaying && curr.state == avtransport.StatePaused {
			log.Info("transport was previously playing, but was paused externally, doing nothing")
			return avtransport.StatePaused, nil
		}

		currentItem, ok := queue.Current()
		if !ok {
			log.Info("reached end of queue, doing nothing")
			return avtransport.StateStopped, nil
		}

		// TODO: maybe check seek?
		if currentItem.HasURI(curr.uri) {
			switch curr.state {
			case avtransport.StatePlaying:
				log.Info("transport URI & state match, doing nothing")
				return desiredState, nil
			default:
				log.Info("transport URI matches, starting playback")
				if err := transport.Play(ctx); err != nil {
					log.WithError(err).Error("could not play transport")
					return desiredState, fmt.Errorf("could not play transport: %w", err)
				}
				return desiredState, nil
			}
		}

		if !transportChanged && currentItem.HasURI(prev.uri) { // TODO: && prev.state == avtransport.StatePlaying ?
			log.Info("controller has fallen behind, skipping track")
			currentItem, ok = queue.Skip()
		}

		seek := prev.elapsed
		if !transportChanged && !currentItem.HasURI(curr.uri) {
			seek = 0
		}

		_, sinks, err := manager.ProtocolInfo(ctx)
		if err != nil {
			log.WithError(err).Error("could not get sink protocols for transport")
			return desiredState, fmt.Errorf("could not get sink protocols for device: %w", err)
		}
		if len(sinks) == 0 {
			log.Error("got 0 sink protocols for transport, expected at least 1")
			return desiredState, errors.New("got 0 sink protocols for device, expected at least 1")
		}
		log.Info("got sink protocols for transport")

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
				log.Info("reached end of queue, doing nothing")
				return avtransport.StateStopped, nil
			}
		}

		log.AddField("new.uri", currentURI)

		// TODO: make this less abrupt, gapless playback, etc.
		if curr.state != avtransport.StateStopped {
			log.Info("temporarily stopping transport")
			_ = transport.Stop(ctx)
		}

		metadata := &upnpav.DIDLLite{Items: []upnpav.Item{currentItem}}
		if err := transport.SetCurrentURI(ctx, currentURI, metadata); err != nil {
			log.WithError(err).Error("could not set transport URI")
			return desiredState, fmt.Errorf("could not set transport URI: %w", err)
		}
		log.Info("set transport URI")

		if err := transport.Play(ctx); err != nil {
			log.WithError(err).Error("could not start transport playing")
			return desiredState, fmt.Errorf("could not play: %w", err)
		}
		log.Info("started transport playing")

		if transportChanged && seek != 0 {
			log.AddField("seek", seek)
			if err := transport.Seek(ctx, seek); err != nil {
				log.WithError(err).Error("could not seek")
				return desiredState, fmt.Errorf("could not seek: %w", err)
			}
			log.Info("seeked transport")
		}
		return desiredState, nil

	default:
		panic(fmt.Sprintf("can only have desired states %v, got %q", []avtransport.State{avtransport.StatePlaying, avtransport.StatePaused, avtransport.StateStopped}, desiredState))
	}
}

func newTransportState(ctx context.Context, transport avtransport.Interface) (transportState, error) {
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

func udnOrDefault(device *upnp.Device, def string) string {
	if device == nil {
		return def
	}
	return device.UDN
}
