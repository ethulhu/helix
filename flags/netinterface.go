package flags

import "net"

func NetInterface(raw string) (interface{}, error) {
	if raw == "" {
		return (*net.Interface)(nil), nil
	}
	return net.InterfaceByName(raw)
}
