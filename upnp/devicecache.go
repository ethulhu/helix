// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
)

type (
	// DeviceCache is an automatically refreshing cache of UPnP devices, addressable by UDN.
	DeviceCache struct {
		urn   ssdp.URN
		iface *net.Interface

		mu      sync.Mutex
		devices map[string]*ssdp.Device
	}

	DeviceCacheOptions struct {
		InitialRefresh time.Duration
		StableRefresh  time.Duration

		Interface *net.Interface
	}
)

const (
	discoveryTimeout = 2 * time.Second
)

// NewDeviceCache returns a DeviceCache searching for the given URN, every refresh period, optionally on a specific network interface.
func NewDeviceCache(urn ssdp.URN, options DeviceCacheOptions) *DeviceCache {
	d := &DeviceCache{
		urn:   urn,
		iface: options.Interface,

		devices: map[string]*ssdp.Device{},
	}

	go d.Refresh()
	go func() {
		for {
			for range time.Tick(options.InitialRefresh) {
				d.Refresh()
				if len(d.Devices()) > 0 {
					break
				}
			}
			for range time.Tick(options.StableRefresh) {
				d.Refresh()
				if len(d.Devices()) == 0 {
					break
				}
			}
		}
	}()

	return d
}

// Refresh forces the DeviceCache to update itself by discovering UPnP devices.
func (d *DeviceCache) Refresh() {
	d.mu.Lock()
	defer d.mu.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), discoveryTimeout)
	devices, _, err := ssdp.Discover(ctx, d.urn, d.iface)
	if err != nil {
		log.Printf("could not find UPnP devices for URN %q: %v", d.urn, err)
		return
	}

	newDevices := map[string]*ssdp.Device{}
	for _, device := range devices {
		newDevices[device.UDN] = device
	}

	d.devices = newDevices
}

// Devices lists all currently known Devices.
func (d *DeviceCache) Devices() []*ssdp.Device {
	d.mu.Lock()
	defer d.mu.Unlock()

	var devices []*ssdp.Device
	for _, device := range d.devices {
		devices = append(devices, device)
	}
	return devices
}

// DeviceByUDN returns the Device with a given UDN.
func (d *DeviceCache) DeviceByUDN(udn string) (*ssdp.Device, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	device, ok := d.devices[udn]
	return device, ok
}
