// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"fmt"
	"strings"
	"time"
)

type (
	Duration struct {
		time.Duration
	}
)

func ParseDuration(raw string) (Duration, error) {

	withoutSubseconds := strings.Split(raw, ".")[0]

	parts := strings.Split(withoutSubseconds, ":")

	var hours, minutes, seconds string
	switch len(parts) {
	case 2:
		hours = "0"
		minutes = parts[0]
		seconds = parts[1]
	case 3:
		hours = parts[0]
		if hours == "" {
			hours = "0"
		}
		minutes = parts[1]
		seconds = parts[2]
	default:
		return Duration{0}, fmt.Errorf("invalid number of parts")
	}

	d, err := time.ParseDuration(fmt.Sprintf("%vh%vm%vs", hours, minutes, seconds))
	return Duration{d}, err
}

func (d Duration) String() string {
	td := d.Duration

	hours := time.Duration(td.Truncate(time.Hour).Hours())
	td = td - (hours * time.Hour)
	minutes := time.Duration(td.Truncate(time.Minute).Minutes())
	td = td - (minutes * time.Minute)
	seconds := time.Duration(td.Truncate(time.Second).Seconds())
	td = td - (seconds * time.Second)
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
func (d *Duration) UnmarshalText(raw []byte) error {
	dd, err := ParseDuration(string(raw))
	if err != nil {
		return err
	}
	*d = dd
	return nil
}
