package internal

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

type (
	Devices struct {
		mu      sync.Mutex
		devices map[string]*ssdp.Device
	}
)

func NewDevices(refresh time.Duration) *Devices {
	d := &Devices{
		devices: map[string]*ssdp.Device{},
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
	newDevices, _, err := ssdp.Discover(ctx, ssdp.All)
	if err != nil {
		log.Printf("could not find UPnP devices: %v", err)
		return
	}
	for _, device := range newDevices {
		d.devices[device.UDN] = device
	}
}

func (d *Devices) AVTransportByUDN(udn string) (avtransport.Client, bool) {
	device, ok := d.DeviceByUDN(udn)
	if !ok {
		return nil, false
	}
	soapClient, ok := device.Client(avtransport.Version1)
	if !ok {
		return nil, false
	}
	return avtransport.NewClient(soapClient), true
}
func (d *Devices) ContentDirectoryByUDN(udn string) (contentdirectory.Client, bool) {
	device, ok := d.DeviceByUDN(udn)
	if !ok {
		return nil, false
	}
	soapClient, ok := device.Client(contentdirectory.Version1)
	if !ok {
		return nil, false
	}
	return contentdirectory.NewClient(soapClient), true
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
		if _, ok := device.Client(urn); ok {
			devices = append(devices, device)
		}
	}
	return devices
}
