// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

/*
Package flag wraps Go's built-in flag package, with the addition of idiomatic custom flags.

Custom Flags

Custom flags are wrappers around Go's built-in string flags, with a parser
func. They can be used to parse custom flag types, or to have custom flag
validators, while keeping the parsing & validation with the flag's definition.

	var (
		urlFlag = flag.Custom("url", "", "url to GET", func(raw string) (interface{}, error) {
			return url.Parse(raw)
		})
		outputFlag = flag.Custom("output", "", "output format", func(raw string) (interface{}, error) {
			if !(raw == "table" || raw == "json") {
				return nil, fmt.Errorf("must be either json or table, got %v", raw)
			}
			return raw, nil
		})
	)

	func main() {
		flag.Parse()
		urlFlag := (*urlFlag).(*url.URL)
		outputFlag := (*outputFlag).(string)
	}

*/
package flag
