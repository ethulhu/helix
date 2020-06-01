// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package flag

import (
	"fmt"
	"strconv"
	"strings"
)

func StringEnum(values ...string) ParseFunc {
	return func(raw string) (interface{}, error) {
		for _, value := range values {
			if value == raw {
				return raw, nil
			}
		}
		return raw, fmt.Errorf("must be one of %q", values)
	}
}

func IntList(raw string) (interface{}, error) {
	var ints []int

	if raw == "" {
		return ints, nil
	}

	for _, raw := range strings.Split(raw, ",") {
		i, err := strconv.Atoi(raw)
		if err != nil {
			return ints, err
		}
		ints = append(ints, i)
	}

	return ints, nil
}
