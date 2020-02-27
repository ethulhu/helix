package internal

import (
	"github.com/ethulhu/helix/upnpav"
)

type (
	TrackSequence interface {
		Current() (upnpav.Item, bool)
		Skip()
	}

	TrackList struct {
		list    []upnpav.Item
		current int
	}
)

func (t *TrackList) Current() (upnpav.Item, bool) {
	if t.current < len(t.list) {
		return t.list[t.current], true
	}
	return upnpav.Item{}, false
}

func (t *TrackList) Skip() {
	if t.current < len(t.list) {
		t.current++
	}
}

func (t *TrackList) AddLast(item upnpav.Item) {
	t.list = append(t.list, item)
}
func (t *TrackList) Clear() {
	t.list = nil
	t.current = 0
}
func (t *TrackList) Items() []upnpav.Item {
	return t.list
}
