package main

import (
	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
)

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
	ID    string `json:"id"`
	Title string `json:"title"`

	ItemClass string `json:"itemClass"`

	// Item fields.
	MIMETypes []string `json:"mimeTypes,omitempty"`

	// Container fields.
	Children []object `json:"children,omitempty"`
}

func objectFromContainer(container upnpav.Container) object {
	return object{
		ID:        string(container.ID),
		Title:     container.Title,
		ItemClass: string(container.Class),
	}
}
func objectFromItem(item upnpav.Item) object {
	var mimetypes []string
	for _, r := range item.Resources {
		if r.ProtocolInfo.Protocol != upnpav.ProtocolHTTP {
			continue
		}
		mimetypes = append(mimetypes, r.ProtocolInfo.ContentFormat)
	}

	return object{
		ID:        string(item.ID),
		Title:     item.Title,
		ItemClass: string(item.Class),

		MIMETypes: mimetypes,
	}
}
