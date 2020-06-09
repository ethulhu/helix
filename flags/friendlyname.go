package flags

import (
	"fmt"
	"os"
)

func FriendlyName(raw string) (interface{}, error) {
	if raw == "" {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("Helix (%s)", hostname), nil
	}
	return raw, nil
}
