// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package scpd

func Must(d Document, err error) Document {
	if err != nil {
		panic(err)
	}
	return d
}