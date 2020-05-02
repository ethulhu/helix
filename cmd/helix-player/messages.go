package main

import (
	"github.com/ethulhu/helix/upnp/ssdp"
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

// ContentDirectory messages.

type directory struct {
	UDN  string `json:"udn"`
	Name string `json:"name"`
}

func directoryFromDevice(device *ssdp.Device) directory {
	return directory{
		UDN:  device.UDN,
		Name: device.Name,
	}
}

type object struct {
	Directory string `json:"directory"`
	ID        string `json:"id"`
	Title     string `json:"title"`

	ItemClass string `json:"itemClass"`

	// Item fields.
	MIMETypes []string `json:"mimetypes,omitempty"`

	// Container fields.
	Children []object `json:"children,omitempty"`
}

func objectFromContainer(udn string, container upnpav.Container) object {
	return object{
		Directory: udn,
		ID:        string(container.ID),
		Title:     container.Title,
		ItemClass: string(container.Class),
	}
}
func objectFromItem(udn string, item upnpav.Item) object {
	var mimetypes []string
	for _, r := range item.Resources {
		if r.ProtocolInfo.Protocol != upnpav.ProtocolHTTP {
			continue
		}
		mimetypes = append(mimetypes, r.ProtocolInfo.ContentFormat)
	}

	return object{
		Directory: udn,
		ID:        string(item.ID),
		Title:     item.Title,
		ItemClass: string(item.Class),

		MIMETypes: mimetypes,
	}
}

// AVTransport messages.

type transport struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	State string `json:"state"`
}

func transportFromDeviceAndInfo(device *ssdp.Device, state avtransport.State) transport {
	return transport{
		ID:   device.UDN,
		Name: device.Name,

		State: humanReadableState(state),
	}
}

// Control Point messages.

type queue struct {
	TransportID   string `json:"transport"`
	TransportName string `json:"transportName,omitempty"`
	State         string `json:"state"`
}

func queueFromControlLoop(cl *controlpoint.ControlLoop) queue {
	transportID := "none"
	transportName := ""
	if t := controlLoop.Transport(); t != nil {
		transportID = t.UDN
		transportName = t.Name
	}

	return queue{
		TransportID:   transportID,
		TransportName: transportName,
		State:         humanReadableState(controlLoop.State()),
	}
}
