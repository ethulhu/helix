// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: CC0-1.0

package flags

import "net"

func NetInterface(raw string) (interface{}, error) {
	if raw == "" {
		return (*net.Interface)(nil), nil
	}
	return net.InterfaceByName(raw)
}