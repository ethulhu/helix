package internal

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

type (
	Devices struct {
		mu    sync.Mutex
		iface *net.Interface

		devices map[string]*ssdp.Device
	}
)

func NewDevices(refresh time.Duration, iface *net.Interface) *Devices {
	d := &Devices{
		devices: map[string]*ssdp.Device{},
		iface:   iface,
	}

	go d.Refresh()
	go func() {
		for _ = range time.Tick(refresh) {
			d.Refresh()
		}
	}()

	return d
}

func (d *Devices) Refresh() {
	d.mu.Lock()
	defer d.mu.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	devices, _, err := ssdp.Discover(ctx, ssdp.All, d.iface)
	if err != nil {
		log.Printf("could not find UPnP devices: %v", err)
		return
	}

	newDevices := map[string]*ssdp.Device{}
	for _, device := range devices {
		newDevices[device.UDN] = device
	}

	d.devices = newDevices

}

func (d *Devices) AVTransportByUDN(udn string) (avtransport.Client, bool) {
	device, ok := d.DeviceByUDN(udn)
	if !ok {
		return nil, false
	}
	client, ok := device.SOAPClient(avtransport.Version1)
	if !ok {
		return nil, false
	}
	return avtransport.NewClient(client), true
}
func (d *Devices) ContentDirectoryByUDN(udn string) (contentdirectory.Client, bool) {
	device, ok := d.DeviceByUDN(udn)
	if !ok {
		return nil, false
	}
	client, ok := device.SOAPClient(contentdirectory.Version1)
	if !ok {
		return nil, false
	}
	return contentdirectory.NewClient(client), true
}

func (d *Devices) DeviceByUDN(udn string) (*ssdp.Device, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	device, ok := d.devices[udn]
	return device, ok
}

func (d *Devices) DevicesByURN(urn ssdp.URN) []*ssdp.Device {
	d.mu.Lock()
	defer d.mu.Unlock()

	var devices []*ssdp.Device
	for _, device := range d.devices {
		if _, ok := device.SOAPClient(urn); ok {
			devices = append(devices, device)
		}
	}
	return devices
}
