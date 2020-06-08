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
		device *upnp.Device
		queue  Queue

		state   avtransport.State
		elapsed time.Duration
	}
	transportState struct {
		state    avtransport.State
		uri      string
		nextURI  string
		elapsed  time.Duration
		duration time.Duration
	}

	action int
)

const (
	doNothing action = iota
	play
	pause
	stop
	seek
	setURI
	setNextURI
	skipTrack
)

func (a action) String() string {
	switch a {
	case doNothing:
		return "doNothing"
	case play:
		return "play"
	case pause:
		return "pause"
	case stop:
		return "stop"
	case seek:
		return "seek"
	case setURI:
		return "setURI"
	case setNextURI:
		return "setNextURI"
	case skipTrack:
		return "skipTrack"
	default:
		panic(fmt.Sprintf("unknown action: %#v", a))
	}
}

func NewLoop() *Loop {
	loop := &Loop{
		state: avtransport.StateStopped,
	}

	ctx := context.Background()
	go func() {
		// We're using UDNs instead of pointer equality for the case.
		var prevDevice *upnp.Device
		prevTransportState, err := newTransportState(ctx, nil)
		if err != nil {
			// We passed a nil transport, so it shouldn't be possible to get errors here.
			panic(fmt.Sprintf("could not get initial transport state: %v", err))
		}
		var protocolInfos []upnpav.ProtocolInfo

		for _ = range time.Tick(1 * time.Second) {
			deviceChanged := udnOrDefault(prevDevice, "") != udnOrDefault(loop.device, "")

			if deviceChanged && prevDevice != nil {
				go func() {
					log, ctx := logger.FromContext(ctx)
					log.AddField("transport.previous.udn", prevDevice.UDN)
					log.AddField("transport.previous.name", prevDevice.Name)

					if err := transport(prevDevice).Stop(ctx); err != nil {
						log.WithError(err).Warning("could not stop previous transport")
						return
					}
					log.Info("stopped previous transport")
				}()
			}

			log, ctx := logger.FromContext(ctx)
			if loop.device == nil {
				if deviceChanged {
					log.Info("no current renderer device")
				}
				continue
			}
			log.AddField("transport.udn", loop.device.UDN)
			log.AddField("transport.name", loop.device.Name)

			if deviceChanged { // && loop.device != nil
				var err error
				_, protocolInfos, err = manager(loop.device).ProtocolInfo(ctx)
				if err != nil {
					loop.device = nil
					log.WithError(err).Error("could not get sink protocols for renderer")
					continue
				}
				if len(protocolInfos) == 0 {
					loop.device = nil
					log.WithError(err).Error("got 0 sink protocols for renderer, expected at least 1")
					continue
				}
				log.Info("got sink protocols for renderer")
			}

			currTransport := transport(loop.device)
			currTransportState, err := newTransportState(ctx, currTransport)
			if err != nil {
				log.WithError(err).Error("could not get transport state")
				continue
			}

			log.AddField("current.state", currTransportState.state)
			if currTransportState.state == avtransport.StatePlaying || currTransportState.state == avtransport.StatePaused {
				log.AddField("current.uri", currTransportState.uri)
			}

			newLoopState, newLoopElapsed, action := tick(loop.queue, protocolInfos, prevTransportState, currTransportState, loop.state, loop.elapsed, deviceChanged)

			if loop.state != newLoopState {
				loop.state = newLoopState
				log.AddField("new.state", newLoopState)
				log.Info("updated desired loop state")
			}
			loop.elapsed = newLoopElapsed

			loop.enact(ctx, protocolInfos, action)

			prevTransportState = currTransportState
			prevDevice = loop.device
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
	return loop.device
}
func (loop *Loop) SetTransport(device *upnp.Device) error {
	if device == nil {
		loop.device = nil
		return nil
	}

	if _, ok := device.SOAPInterface(avtransport.Version1); !ok {
		return errors.New("device does not support AVTransport")
	}
	if _, ok := device.SOAPInterface(connectionmanager.Version1); !ok {
		return errors.New("device does not support ConnectionManager")
	}

	loop.device = device
	return nil
}

func (loop *Loop) enact(ctx context.Context, protocolInfos []upnpav.ProtocolInfo, action action) {
	log, ctx := logger.FromContext(ctx)
	transport := transport(loop.device)
	log.AddField("action", action)

	switch action {
	case doNothing:
		log.Debug("doing nothing")

	case skipTrack:
		loop.queue.Skip()
		log.Info("skipped track")

	case play:
		if err := transport.Play(ctx); err != nil {
			log.WithError(err).Warning("could not play transport")
			return
		}
		log.Info("set transport playing")

	case pause:
		if err := transport.Pause(ctx); err != nil {
			log.WithError(err).Warning("could not pause transport")
			return
		}
		log.Info("paused transport")

	case stop:
		if err := transport.Stop(ctx); err != nil {
			log.WithError(err).Warning("could not stop transport")
			return
		}
		log.Info("stopped transport")

	case seek:
		log.AddField("seek", loop.elapsed)
		if err := transport.Seek(ctx, loop.elapsed); err != nil {
			log.WithError(err).Warning("could not seek transport")
			return
		}
		log.Info("seeked transport")

	case setURI:
		item, ok := loop.queue.Current()
		if !ok {
			panic("got empty queue for action setURI")
		}
		uri, ok := item.URIForProtocolInfos(protocolInfos)
		if !ok {
			panic("got an unplayable item for action setURI")
		}

		log.AddField("uri", uri)
		metadata := &upnpav.DIDLLite{Items: []upnpav.Item{item}}

		_ = transport.Stop(ctx)
		if err := transport.SetCurrentURI(ctx, uri, metadata); err != nil {
			log.WithError(err).Error("could not set transport URI")
			return
		}
		if err := transport.Play(ctx); err != nil {
			log.WithError(err).Error("could not start transport playing")
			return
		}
		log.Info("set transport URI")

	case setNextURI:
		item, ok := loop.queue.Next()
		if !ok {
			panic("got empty queue for action setNextURI")
		}
		uri, ok := item.URIForProtocolInfos(protocolInfos)
		if !ok {
			panic("got an unplayable item for action setNextURI")
		}

		log.AddField("next.uri", uri)
		metadata := &upnpav.DIDLLite{Items: []upnpav.Item{item}}

		if err := transport.SetNextURI(ctx, uri, metadata); err != nil {
			log.WithError(err).Warning("could not set next transport URI")
		}
		log.Info("set next transport URI")

	default:
		panic(fmt.Sprintf("got unhandled action %#v", action))
	}
}

// manager will panic if device is invalid because SetTransport should make that impossible.
func manager(device *upnp.Device) connectionmanager.Interface {
	managerClient, ok := device.SOAPInterface(connectionmanager.Version1)
	if !ok {
		panic(fmt.Sprintf("transport does not support ConnectionManager"))
	}
	return connectionmanager.NewClient(managerClient)
}

// transport will panic if device is invalid because SetTransport should make that impossible.
func transport(device *upnp.Device) avtransport.Interface {
	transportClient, ok := device.SOAPInterface(avtransport.Version1)
	if !ok {
		panic(fmt.Sprintf("transport does not support AVTransport"))
	}
	return avtransport.NewClient(transportClient)
}

// tick is a 7-argument monstrosity to make it clear what it consumes.
func tick(
	queue Queue,
	protocolInfos []upnpav.ProtocolInfo,
	prev transportState,
	curr transportState,
	loopState avtransport.State,
	loopElapsed time.Duration,
	deviceChanged bool) (avtransport.State, time.Duration, action) {

	if curr.state == avtransport.StateTransitioning {
		return loopState, loopElapsed, doNothing
	}

	if queue == nil {
		return avtransport.StateStopped, 0, doNothing
	}

	switch loopState {

	case avtransport.StateStopped:
		switch curr.state {
		case avtransport.StateStopped:
			return avtransport.StateStopped, 0, doNothing
		default:
			return avtransport.StateStopped, 0, stop
		}

	case avtransport.StatePaused:
		switch curr.state {
		case avtransport.StatePaused:
			return avtransport.StatePaused, curr.elapsed, doNothing
		default:
			// TODO: also check URI & elapsed time?
			if prev.state == avtransport.StatePaused && curr.state == avtransport.StatePlaying {
				return avtransport.StatePlaying, curr.elapsed, doNothing
			}

			return avtransport.StatePaused, curr.elapsed, pause
		}

	case avtransport.StatePlaying:
		// TODO: generalize to "everything else matches, and prev.state matched, but the curr.state differs"?
		if prev.state == avtransport.StatePlaying && curr.state == avtransport.StatePaused {
			return avtransport.StatePaused, curr.elapsed, doNothing
		}

		currentItem, tracksLeftInQueue := queue.Current()
		if !tracksLeftInQueue {
			return avtransport.StateStopped, 0, stop
		}
		if _, ok := currentItem.URIForProtocolInfos(protocolInfos); !ok {
			return avtransport.StatePlaying, 0, skipTrack
		}

		if deviceChanged {
			return avtransport.StatePlaying, prev.elapsed, setURI
		}

		if !currentItem.HasURI(curr.uri) {
			if currentItem.HasURI(prev.uri) {
				return avtransport.StatePlaying, 0, skipTrack
			}
			return avtransport.StatePlaying, 0, setURI
		}
		if sufficientlyDeviant(curr.elapsed, loopElapsed) {
			return avtransport.StatePlaying, loopElapsed, seek
		}
		if curr.state != avtransport.StatePlaying {
			return avtransport.StatePlaying, curr.elapsed, play
		}

		if nextItem, ok := queue.Next(); ok {
			if nextURI, ok := nextItem.URIForProtocolInfos(protocolInfos); ok {
				if curr.nextURI != nextURI {
					return avtransport.StatePlaying, curr.elapsed, setNextURI
				}
			}
		}

		return avtransport.StatePlaying, curr.elapsed, doNothing

	default:
		panic(fmt.Sprintf("can only have loop states %v, got %q", []avtransport.State{avtransport.StatePlaying, avtransport.StatePaused, avtransport.StateStopped}, loopState))
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
		uri, _, duration, elapsed, err := transport.PositionInfo(ctx)
		if err != nil {
			return t, nil
		}
		t.uri = uri
		t.elapsed = elapsed
		t.duration = duration

		_, _, nextURI, _, err := transport.MediaInfo(ctx)
		if err != nil {
			return t, nil
		}
		t.nextURI = nextURI
	}
	return t, nil
}

func udnOrDefault(device *upnp.Device, def string) string {
	if device == nil {
		return def
	}
	return device.UDN
}

func sufficientlyDeviant(t1, t2 time.Duration) bool {
	diff := t1 - t2
	if diff < 0 {
		diff = -diff
	}
	return diff > 2*time.Second
}
