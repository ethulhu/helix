// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package flag

import (
	"fmt"
	"os"
	"time"
)

var (
	CommandLine = NewFlagSet(os.Args[0], ExitOnError)
	Usage       = func() {
		fmt.Fprintf(CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		CommandLine.PrintDefaults()
	}
)

func init() {
	CommandLine.Usage = runUsageVariable
}
func runUsageVariable() {
	Usage()
}

func Parse() {
	_ = CommandLine.Parse(os.Args[1:])
}

func String(flagName, defaultValue, description string) *string {
	return CommandLine.String(flagName, defaultValue, description)
}
func Int(flagName string, defaultValue int, description string) *int {
	return CommandLine.Int(flagName, defaultValue, description)
}
func Duration(flagName string, defaultValue time.Duration, description string) *time.Duration {
	return CommandLine.Duration(flagName, defaultValue, description)
}
func Custom(flagName, defaultValue, description string, parser ParseFunc) *interface{} {
	return CommandLine.Custom(flagName, defaultValue, description, parser)
}
func Bool(flagName string, defaultValue bool, description string) *bool {
	return CommandLine.Bool(flagName, defaultValue, description)
}
