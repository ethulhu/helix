// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/controlpoint"
)

func humanReadableState(state avtransport.State) string {
	switch state {
	case avtransport.StatePlaying:
		return "playing"
	case avtransport.StatePaused:
		return "paused"
	case avtransport.StateStopped:
		return "stopped"
	default:
		return string(state)
	}
}

type objectMetadata struct {
	Title     string `json:"title"`
	ItemClass string `json:"itemClass"`

	// Item fields.
	MIMETypes []string `json:"mimetypes,omitempty"`
}

func objectMetadataFromItem(item upnpav.Item) objectMetadata {
	var mimetypes []string
	for _, r := range item.Resources {
		if r.ProtocolInfo.Protocol != upnpav.ProtocolHTTP {
			continue
		}
		mimetypes = append(mimetypes, r.ProtocolInfo.ContentFormat)
	}

	return objectMetadata{
		Title:     item.Title,
		ItemClass: string(item.Class),
		MIMETypes: mimetypes,
	}
}
func objectMetadataFromContainer(container upnpav.Container) objectMetadata {
	return objectMetadata{
		Title:     container.Title,
		ItemClass: string(container.Class),
	}
}

// ContentDirectory messages.

type directory struct {
	UDN  string `json:"udn"`
	Name string `json:"name"`
}

func directoryFromDevice(device *upnp.Device) directory {
	return directory{
		UDN:  device.UDN,
		Name: device.Name,
	}
}

type directoryObject struct {
	Directory string `json:"directory"`
	ID        string `json:"id"`

	objectMetadata

	// Container fields.
	Children []directoryObject `json:"children,omitempty"`
}

func directoryObjectFromContainer(udn string, container upnpav.Container) directoryObject {
	return directoryObject{
		Directory:      udn,
		ID:             string(container.ID),
		objectMetadata: objectMetadataFromContainer(container),
	}
}
func directoryObjectFromItem(udn string, item upnpav.Item) directoryObject {
	return directoryObject{
		Directory:      udn,
		ID:             string(item.ID),
		objectMetadata: objectMetadataFromItem(item),
	}
}

// AVTransport messages.

type transport struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	State string `json:"state"`
}

func transportFromDeviceAndInfo(device *upnp.Device, state avtransport.State) transport {
	return transport{
		ID:   device.UDN,
		Name: device.Name,

		State: humanReadableState(state),
	}
}

// Control Point messages.

type controlPoint struct {
	TransportID   string  `json:"transport"`
	TransportName string  `json:"transportName,omitempty"`
	State         string  `json:"state"`
	Elapsed       float64 `json:"elapsedSeconds,omitempty"`
	Duration      float64 `json:"durationSeconds,omitempty"`
}

func controlPointFromLoop(cl *controlpoint.Loop) controlPoint {
	transportID := "none"
	transportName := ""
	if t := cl.Transport(); t != nil {
		transportID = t.UDN
		transportName = t.Name
	}

	return controlPoint{
		TransportID:   transportID,
		TransportName: transportName,
		State:         humanReadableState(cl.State()),
		Elapsed:       float64(cl.Elapsed().Seconds()),
		Duration:      float64(cl.Duration().Seconds()),
	}
}

type queue struct {
	Upcoming []queueItem `json:"upcoming"`
	History  []queueItem `json:"history"`
}

func queueFromTrackList(tl *controlpoint.TrackList) queue {
	upcoming := []queueItem{}
	for _, qi := range tl.Upcoming() {
		upcoming = append(upcoming, queueItemFromQueueItem(qi))
	}

	history := []queueItem{}
	for _, qi := range tl.History() {
		history = append(history, queueItemFromQueueItem(qi))
	}

	return queue{
		Upcoming: upcoming,
		History:  history,
	}
}

type queueItem struct {
	ID int `json:"id"`

	objectMetadata
}

func queueItemFromQueueItem(qi controlpoint.QueueItem) queueItem {
	return queueItem{
		ID:             qi.ID,
		objectMetadata: objectMetadataFromItem(qi.Item),
	}
}
