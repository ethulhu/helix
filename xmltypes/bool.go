// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package xmltypes

import "fmt"

type (
	// IntBool is a bool that should be serialized as either "1" or "0".
	IntBool bool

	// YesNoBool is a bool that should be serialized as either "yes" or "no".
	YesNoBool bool
)

const (
	Yes = YesNoBool(true)
	No  = YesNoBool(false)
)

func (i IntBool) MarshalText() ([]byte, error) {
	if bool(i) {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}
func (i *IntBool) UnmarshalText(raw []byte) error {
	b, err := unmarshalBool(raw)
	if err != nil {
		return err
	}
	*i = IntBool(b)
	return nil
}

func (y YesNoBool) MarshalText() ([]byte, error) {
	if bool(y) {
		return []byte("yes"), nil
	}
	return []byte("no"), nil
}
func (y *YesNoBool) UnmarshalText(raw []byte) error {
	b, err := unmarshalBool(raw)
	if err != nil {
		return err
	}
	*y = YesNoBool(b)
	return nil
}

func unmarshalBool(raw []byte) (bool, error) {
	switch string(raw) {
	case "1":
		fallthrough
	case "true":
		fallthrough
	case "yes":
		return true, nil

	case "0":
		fallthrough
	case "false":
		fallthrough
	case "no":
		return false, nil

	default:
		return false, fmt.Errorf("must be one of [0,1,true,false,yes,no], got %s", raw)
	}
}
