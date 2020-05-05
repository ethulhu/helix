package controlpoint

import (
	"reflect"
	"testing"

	"github.com/ethulhu/helix/upnpav"
)

func TestTrackListAppendUpcoming(t *testing.T) {
	tracks := []upnpav.Item{
		{Title: "a"},
		{Title: "b"},
		{Title: "c"},
	}

	tl := NewTrackList()
	for _, track := range tracks {
		tl.Append(track)
	}

	for i, qi := range tl.Upcoming() {
		if !reflect.DeepEqual(qi.Item, tracks[i]) {
			t.Errorf("tl.Upcoming()[%d] == %+v, wanted %+v", i, qi.Item, tracks[i])
		}
	}
}

func TestTrackListAppendHistory(t *testing.T) {
	tracks := []upnpav.Item{
		{Title: "a"},
		{Title: "b"},
		{Title: "c"},
	}

	tl := NewTrackList()
	for _, track := range tracks {
		tl.Append(track)
	}

	for range tracks {
		_, _ = tl.Skip()
	}

	for i, qi := range tl.History() {
		if !reflect.DeepEqual(qi.Item, tracks[i]) {
			t.Errorf("tl.Upcoming()[%d] == %+v, wanted %+v", i, qi.Item, tracks[i])
		}
	}
}
