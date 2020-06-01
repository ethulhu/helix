// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"encoding/xml"
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

func (d Duration) MarshalXML(enc *xml.Encoder, el xml.StartElement) error {
	return enc.EncodeElement(d.String(), el)
}
func (d *Duration) UnmarshalXML(dec *xml.Decoder, el xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &el); err != nil {
		return err
	}

	newD, err := ParseDuration(s)
	if err != nil {
		return err
	}

	*d = newD
	return nil
}
