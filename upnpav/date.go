// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
	"encoding/xml"
	"time"
)

type (
	Date struct {
		time.Time
	}
)

func ParseDate(raw string) (Date, error) {
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", raw)
		if err != nil {
			return Date{}, err
		}
	}
	return Date{t}, nil
}

func (d Date) String() string {
	return d.Format("2006-01-02")
}

func (d Date) MarshalXML(enc *xml.Encoder, el xml.StartElement) error {
	return enc.EncodeElement(d.String(), el)
}

func (d *Date) UnmarshalXML(dec *xml.Decoder, el xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &el); err != nil {
		return err
	}

	newD, err := ParseDate(s)
	if err != nil {
		return err
	}

	*d = newD
	return nil
}
