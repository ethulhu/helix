// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseResolution(raw string) (*Resolution, error) {
	parts := strings.Split(raw, "x")
	if len(parts) != 2 {
		return nil, fmt.Errorf("expected 2 dimensions, got %d", len(parts))
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("could not parse width: %v", err)
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse height: %v", err)
	}

	return &Resolution{Width: width, Height: height}, nil
}
func (r *Resolution) String() string {
	return fmt.Sprintf("%dx%d", r.Width, r.Height)
}
