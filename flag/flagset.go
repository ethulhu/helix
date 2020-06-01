// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package flag

import (
	"flag"
	"fmt"
	"os"
)

type (
	ErrorHandling = flag.ErrorHandling

	ParseFunc func(string) (interface{}, error)

	FlagSet struct {
		flag.FlagSet

		customFlags []func() error
	}
)

const (
	ContinueOnError = flag.ContinueOnError
	ExitOnError     = flag.ExitOnError
	PanicOnError    = flag.PanicOnError
)

func (f *FlagSet) Parse(arguments []string) error {
	if err := f.FlagSet.Parse(arguments); err != nil {
		return err
	}

	for _, customFlag := range f.customFlags {
		if err := customFlag(); err != nil {
			switch f.FlagSet.ErrorHandling() {
			case flag.ContinueOnError:
				return err
			case flag.ExitOnError:
				fmt.Fprintf(os.Stdout, "%v\n\n", err)
				f.Usage()
				os.Exit(2)
			case flag.PanicOnError:
				panic(err)
			}
		}
	}
	return nil
}

func NewFlagSet(name string, handling ErrorHandling) *FlagSet {
	return &FlagSet{
		FlagSet: *flag.NewFlagSet(name, handling),
	}
}

func (f *FlagSet) Custom(flagName, defaultValue, description string, parser ParseFunc) *interface{} {
	rawFlag := f.String(flagName, defaultValue, description)

	var value interface{}

	f.customFlags = append(f.customFlags, func() error {
		var err error
		value, err = parser(*rawFlag)
		if err != nil {
			return fmt.Errorf("invalid value %q for flag -%s: %w", *rawFlag, flagName, err)
		}
		return nil
	})

	return &value
}
