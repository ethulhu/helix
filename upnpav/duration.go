// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type (
	// Duration is of the form H+:MM:SS[.F+] or H+:MM:SS[.F0/F1], where:
	// H+ is 0 or more digits for hours,
	// MM is exactly 2 digits for minutes,
	// SS is exactly 2 digits for seconds,
	// F+ is 0 or more digits for fractional seconds,
	// F0/F1 is a fraction, F0 & F1 are at least 1 digit, and F0/F1 < 1.
	Duration struct {
		time.Duration
	}
)

func ParseDuration(raw string) (Duration, error) {
	var withoutSubseconds, subseconds string
	switch parts := strings.Split(raw, "."); len(parts) {
	case 1:
		withoutSubseconds = parts[0]
	case 2:
		withoutSubseconds = parts[0]
		subseconds = parts[1]
	default:
		return Duration{0}, fmt.Errorf("invalid duration format")
	}

	var hours, minutes, seconds string
	switch parts := strings.Split(withoutSubseconds, ":"); len(parts) {
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

	milliseconds := 0
	if subseconds != "" {
		fractions := strings.Split(subseconds, "/")
		switch len(fractions) {
		case 1:
			fraction, err := strconv.ParseFloat("0."+subseconds, 64)
			if err != nil {
				return Duration{0}, fmt.Errorf("could not parse milliseconds %q: %v", subseconds, err)
			}
			milliseconds = int(fraction * 1000)
		case 2:
			top, err := strconv.Atoi(fractions[0])
			if err != nil {
				return Duration{0}, fmt.Errorf("could not parse top of fraction %q: %v", fractions[0], err)
			}
			bottom, err := strconv.Atoi(fractions[1])
			if err != nil {
				return Duration{0}, fmt.Errorf("could not parse bottom of fraction %q: %v", fractions[1], err)
			}
			milliseconds = (1000 * top) / bottom
		default:
			return Duration{0}, fmt.Errorf("invalid subseconds")
		}
	}

	d, err := time.ParseDuration(fmt.Sprintf("%vh%vm%vs%vms", hours, minutes, seconds, milliseconds))
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

	if td == 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}

	milliseconds := time.Duration(td.Truncate(time.Millisecond).Milliseconds())
	return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
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
