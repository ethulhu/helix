// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"fmt"
	"strings"
	"time"
)

func ParseDuration(raw string) (time.Duration, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid number of parts")
	}

	return time.ParseDuration(fmt.Sprintf("%vh%vm%vs", parts[0], parts[1], parts[2]))
}

func FormatDuration(d time.Duration) string {
	hours := time.Duration(d.Truncate(time.Hour).Hours())
	d = d - (hours * time.Hour)
	minutes := time.Duration(d.Truncate(time.Minute).Minutes())
	d = d - (minutes * time.Minute)
	seconds := time.Duration(d.Truncate(time.Second).Seconds())
	d = d - (seconds * time.Second)
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}
