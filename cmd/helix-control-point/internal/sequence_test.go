package internal

import (
	"testing"

	"github.com/ethulhu/helix/upnpav"
)

func TestTrackList(t *testing.T) {
	tl := &TrackList{}

	tl.AddLast(upnpav.Item{Title: "mew"})
	tl.AddLast(upnpav.Item{Title: "purr"})

	current, ok := tl.Current()
	if !ok {
		t.Errorf("[1]: expected track \"mew\", got nothing")
	}
	if current.Title != "mew" {
		if !ok {
			t.Errorf("[1]: expected track \"mew\", got %v", current)
		}
	}

	tl.Skip()

	current, ok = tl.Current()
	if !ok {
		t.Errorf("[2]: expected track \"purr\", got nothing")
	}
	if current.Title != "purr" {
		if !ok {
			t.Errorf("[2]: expected track \"purr\", got %v", current)
		}
	}

	tl.Skip()

	current, ok = tl.Current()
	if ok {
		t.Errorf("[3]: expected nothing, got %v", current)
	}
}
