package xmltypes

import (
	"strconv"
	"strings"
)

type (
	CommaSeparatedStrings []string
	CommaSeparatedInts    []int
)

func (css CommaSeparatedStrings) MarshalText() ([]byte, error) {
	return []byte(strings.Join([]string(css), ",")), nil
}
func (css *CommaSeparatedStrings) UnmarshalText(raw []byte) error {
	if len(raw) == 0 {
		*css = nil
		return nil
	}
	*css = CommaSeparatedStrings(strings.Split(string(raw), ","))
	return nil
}

func (csi CommaSeparatedInts) MarshalText() ([]byte, error) {
	var strs []string
	for _, i := range csi {
		strs = append(strs, strconv.Itoa(i))
	}
	return []byte(strings.Join(strs, ",")), nil
}
func (csi *CommaSeparatedInts) UnmarshalText(raw []byte) error {
	if len(raw) == 0 {
		*csi = nil
		return nil
	}

	var ints []int
	for _, str := range strings.Split(string(raw), ",") {
		i, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		ints = append(ints, i)
	}
	*csi = CommaSeparatedInts(ints)
	return nil
}
