// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

import (
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

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(raw []byte) error {
	dd, err := ParseDate(string(raw))
	if err != nil {
		return err
	}
	*d = dd
	return nil
}
