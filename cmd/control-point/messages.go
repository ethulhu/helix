package main

import (
	"github.com/ethulhu/helix/upnpav/avtransport"
)

type (
	transportStatus struct {
		UDN     string            `json:"udn"`
		Name    string            `json:"name"`
		State   avtransport.State `json:"state"`
		Playing *playingStatus    `json:"playing,omitempty"`
	}
	playingStatus struct {
		Name     string `json:"name"`
		Duration int64  `json:"duration_ms"`
		Elapsed  int64  `json:"elapsed_ms"`
	}
)
