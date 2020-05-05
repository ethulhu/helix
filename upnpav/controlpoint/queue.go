package controlpoint

import (
	"github.com/ethulhu/helix/upnpav"
)

type (
	Queue interface {
		Skip() (upnpav.Item, bool)
		Current() (upnpav.Item, bool)
	}
	TrackList struct {
		items   []upnpav.Item
		current int
	}
)

func NewTrackList() *TrackList {
	return &TrackList{}
}

func (t *TrackList) Items() []upnpav.Item {
	return t.items
}
func (t *TrackList) Append(item upnpav.Item) {
	t.items = append(t.items, item)
}

func (t *TrackList) Skip() (upnpav.Item, bool) {
	if t.current < len(t.items) {
		t.current++
	}
	return t.Current()
}
func (t *TrackList) Current() (upnpav.Item, bool) {
	if t.current < len(t.items) {
		return t.items[t.current], true
	}
	return upnpav.Item{}, false
}
func (t *TrackList) RemoveAll() {
	t.items = nil
	t.current = 0
}
