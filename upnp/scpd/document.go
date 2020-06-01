// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package scpd

import (
	"encoding/xml"
	"fmt"

	"github.com/ethulhu/helix/xmltypes"
)

const xmlns = "urn:schemas-upnp-org:service-1-0"

var Version = SpecVersion{
	Major: 1,
	Minor: 0,
}

type (
	Document struct {
		XMLName        xml.Name        `xml:"urn:schemas-upnp-org:service-1-0 scpd"`
		SpecVersion    SpecVersion     `xml:"specVersion"`
		StateVariables []StateVariable `xml:"serviceStateTable>stateVariable"`
		Actions        []Action        `xml:"actionList>action"`
	}

	SpecVersion struct {
		Major int `xml:"major"`
		Minor int `xml:"minor"`
	}

	StateVariable struct {
		Name                string             `xml:"name"`
		SendEventsAttribute xmltypes.YesNoBool `xml:"sendEventsAttribute"`
		DataType            string             `xml:"dataType"`
		AllowedValues       *AllowedValues     `xml:"allowedValueList,omitempty"`
		AllowedValueRange   *AllowedValueRange `xml:"allowedValueRange,omitempty"`
	}

	AllowedValues struct {
		Values []string `xml:"allowedValues"`
	}
	AllowedValueRange struct {
		Minimum int `xml:"minimum"`
		Maximum int `xml:"maximum,omitempty"`
		Step    int `xml:"step,omitempty"`
	}

	Action struct {
		Name      string     `xml:"name"`
		Arguments []Argument `xml:"argumentList>argument"`
	}

	Argument struct {
		Name                 string    `xml:"name"`
		Direction            Direction `xml:"direction"`
		RelatedStateVariable string    `xml:"relatedStateVariable"`
	}

	Direction int
)

const (
	Unknown Direction = iota
	In
	Out
)

func (d Direction) MarshalText() ([]byte, error) {
	switch d {
	case In:
		return []byte("in"), nil
	case Out:
		return []byte("out"), nil
	default:
		return nil, fmt.Errorf("direction must be In or Out, found %v", d)
	}
}
func (d *Direction) UnmarshalText(raw []byte) error {
	switch string(raw) {
	case "in":
		*d = In
		return nil
	case "out":
		*d = Out
		return nil
	default:
		return fmt.Errorf("invalid direction: %s", raw)
	}
}
