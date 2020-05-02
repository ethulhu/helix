package controlpoint

import (
	"github.com/ethulhu/helix/upnpav"
)

type (
	Queue interface {
		Skip()
		Current() (upnpav.Item, bool)
	}
	TrackList struct {
		items   []upnpav.Item
		current int
	}
)

func NewTrackList() *TrackList{
	return &TrackList{}
}

func (t *TrackList) Items() []upnpav.Item{
	return t.items
}
func (t *TrackList) Append(item upnpav.Item) {
	t.items = append(t.items, item)
}

func (t *TrackList) Skip() {
	if t.current < len(t.items) {
		t.current++
	}
}
func (t *TrackList) Current() (upnpav.Item, bool) {
	if t.current < len(t.items) {
		return t.items[t.current], true
	}
	return upnpav.Item{}, false
}
