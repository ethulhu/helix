// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package flags

import (
	"crypto/rand"
	"fmt"
)

func UDN(raw string) (interface{}, error) {
	if raw == "" {
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err != nil {
			panic(err)
		}
		return fmt.Sprintf("uuid:%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:]), nil
	}
	return raw, nil
}
