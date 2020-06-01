// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package flag

import (
	"testing"
	"time"
)

func TestFlagSetCustom(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)

	i := fs.Int("i", 0, "i")
	d := fs.Custom("d", "1s", "d", func(raw string) (interface{}, error) {
		return time.ParseDuration(raw)
	})
	e := fs.Custom("e", "", "e", StringEnum("json", "table"))

	if err := fs.Parse([]string{"-i", "12", "-d", "3m", "-e", "json"}); err != nil {
		t.Fatalf("got error: %v", err)
	}

	if *i != 12 {
		t.Errorf("-i == %v, wanted %v", *i, 12)
	}
	if (*d).(time.Duration) != 3*time.Minute {
		t.Errorf("-d == %v, wanted %v", *d, 3*time.Minute)
	}
	if (*e).(string) != "json" {
		t.Errorf("-e == %v, wanted %v", *e, "json")
	}
}
