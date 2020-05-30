package scpd

import (
	"encoding/xml"
	"fmt"
)

const xmlns = "urn:schemas-upnp-org:service-1-0"

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
		SendEventsAttribute Bool               `xml:"sendEventsAttribute"`
		DataType            string             `xml:"dataType"`
		AllowedValues       *AllowedValues     `xml:"allowedValueList,omitempty"`
		AllowedValueRange   *AllowedValueRange `xml:"allowedValueRange,omitempty"`
	}

	Bool bool

	AllowedValues struct {
		Values []string `xml:"allowedValues"`
	}
	AllowedValueRange struct {
		Minimum int `xml:"minimum"`
		Maximum int `xml:"maximum"`
		Step    int `xml:"step"`
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

	Direction bool
)

const (
	In  = Direction(true)
	Out = Direction(false)
)

func (d Direction) MarshalXML(e *xml.Encoder, el xml.StartElement) error {
	s := "out"
	if d == In {
		s = "in"
	}
	return e.EncodeElement(s, el)
}
func (d *Direction) UnmarshalXML(dec *xml.Decoder, el xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &el); err != nil {
		return err
	}
	switch s {
	case "in":
		*d = In
	case "out":
		*d = Out
	default:
		return fmt.Errorf("invalid direction: %v", s)
	}
	return nil
}

const (
	Yes = Bool(true)
	No  = Bool(false)
)

func (b Bool) MarshalXML(e *xml.Encoder, el xml.StartElement) error {
	s := "yes"
	if b == No {
		s = "no"
	}
	return e.EncodeElement(s, el)
}
func (b *Bool) UnmarshalXML(d *xml.Decoder, el xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &el); err != nil {
		return err
	}

	switch s {
	case "1":
		fallthrough
	case "true":
		fallthrough
	case "yes":
		*b = Yes

	case "0":
		fallthrough
	case "false":
		fallthrough
	case "no":
		*b = No

	default:
		return fmt.Errorf("invalid boolean: %v", s)
	}
	return nil
}
