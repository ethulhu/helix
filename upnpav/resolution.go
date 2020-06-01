// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	// Resolution of the resource of the form [0-9]+x[0-9]+, e.g. 4x2.
	Resolution struct {
		Height, Width int
	}
)

func ParseResolution(raw string) (Resolution, error) {
	parts := strings.Split(raw, "x")
	if len(parts) != 2 {
		return Resolution{}, fmt.Errorf("expected 2 dimensions, got %d", len(parts))
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return Resolution{}, fmt.Errorf("could not parse width: %v", err)
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return Resolution{}, fmt.Errorf("could not parse height: %v", err)
	}

	return Resolution{Width: width, Height: height}, nil
}

func (r Resolution) String() string {
	return fmt.Sprintf("%dx%d", r.Width, r.Height)
}

func (r Resolution) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}
func (r *Resolution) UnmarshalText(raw []byte) error {
	rr, err := ParseResolution(string(raw))
	if err != nil {
		return err
	}
	*r = rr
	return nil
}
