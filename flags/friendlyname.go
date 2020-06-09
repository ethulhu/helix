// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: CC0-1.0

package flags

import (
	"fmt"
	"os"
)

func FriendlyName(raw string) (interface{}, error) {
	if raw == "" {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("Helix (%s)", hostname), nil
	}
	return raw, nil
}