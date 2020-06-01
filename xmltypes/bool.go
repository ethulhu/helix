// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package xmltypes

import "fmt"

type (
	// IntBool is a bool that should be serialized as either "1" or "0".
	IntBool bool
)

func (i IntBool) MarshalText() ([]byte, error) {
	if bool(i) {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

func (i *IntBool) UnmarshalText(raw []byte) error {
	switch string(raw) {
	case "0":
		fallthrough
	case "false":
		fallthrough
	case "no":
		*i = IntBool(false)
		return nil

	case "1":
		fallthrough
	case "true":
		fallthrough
	case "yes":
		*i = IntBool(true)
		return nil

	default:
		return fmt.Errorf("must be one of [0,1,true,false,yes,no], got %s", raw)
	}
}
