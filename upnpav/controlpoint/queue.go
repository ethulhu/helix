// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package controlpoint

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/ethulhu/helix/upnpav"
)

type (
	Queue interface {
		Skip()
		Current() (upnpav.Item, bool)
		Next() (upnpav.Item, bool)
	}
	TrackList struct {
		items   map[int]upnpav.Item
		order   []int
		current int

		mu sync.Mutex
	}

	QueueItem struct {
		ID   int
		Item upnpav.Item
	}
)

func NewTrackList() *TrackList {
	return &TrackList{
		items: map[int]upnpav.Item{},
	}
}

func (t *TrackList) Items() []QueueItem {
	t.mu.Lock()
	defer t.mu.Unlock()

	var queueItems []QueueItem
	for id, item := range t.items {
		queueItems = append(queueItems, QueueItem{id, item})
	}
	return queueItems
}
func (t *TrackList) Upcoming() []QueueItem {
	t.mu.Lock()
	defer t.mu.Unlock()

	var queueItems []QueueItem
	for _, id := range t.order[t.current:] {
		queueItems = append(queueItems, QueueItem{id, t.items[id]})
	}
	return queueItems
}
func (t *TrackList) History() []QueueItem {
	t.mu.Lock()
	defer t.mu.Unlock()

	var queueItems []QueueItem
	for _, id := range t.order[:t.current] {
		queueItems = append(queueItems, QueueItem{id, t.items[id]})
	}
	return queueItems
}

func (t *TrackList) Append(item upnpav.Item) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	var id int
	for {
		id = int(rand.Int31())
		// id MUST NOT be 0.
		if _, alreadyExists := t.items[id]; !alreadyExists && id != 0 {
			break
		}
	}
	t.items[id] = item
	t.order = append(t.order, id)

	return id
}
func (t *TrackList) SetCurrent(id int) error {
	for i := range t.order {
		if t.order[i] == id {
			t.current = i
			return nil
		}
	}
	return errors.New("unknown id")
}
func (t *TrackList) Remove(id int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.items[id]; !exists {
		return
	}

	for i := range t.order {
		if t.order[i] == id {
			if i < t.current {
				t.current--
			}
			t.order = append(t.order[:i], t.order[i+1:]...)
			break
		}
	}
	delete(t.items, id)
}
func (t *TrackList) RemoveAll() {
	t.items = nil
	t.order = nil
	t.current = 0
}

func (t *TrackList) Skip() {
	if t.current < len(t.order) {
		t.current++
	}
}
func (t *TrackList) Current() (upnpav.Item, bool) {
	return t.atIndex(t.current)
}
func (t *TrackList) Next() (upnpav.Item, bool) {
	return t.atIndex(t.current + 1)
}
func (t *TrackList) atIndex(i int) (upnpav.Item, bool) {
	if i < len(t.order) {
		return t.items[t.order[i]], true
	}
	return upnpav.Item{}, false
}
