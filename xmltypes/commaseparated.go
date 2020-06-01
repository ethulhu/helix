package xmltypes

import "strings"

type (
	CommaSeparatedStrings []string
)

func (css CommaSeparatedStrings) MarshalText() ([]byte, error) {
	return []byte(strings.Join([]string(css), ",")), nil
}
func (css *CommaSeparatedStrings) UnmarshalText(raw []byte) error {
	if len(raw) == 0 {
		*css = nil
	}
	*css = CommaSeparatedStrings(strings.Split(string(raw), ","))
	return nil
}
