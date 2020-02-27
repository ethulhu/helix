package upnpav

import (
	"encoding/xml"
	"fmt"
)

type (
	DIDL struct {
		XMLName    struct{}    `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ DIDL-Lite"`
		Containers []Container `xml:"container"`
		Items      []Item      `xml:"item"`
	}
	Container struct {
		ID       Object `xml:"id,attr"`
		ParentID Object `xml:"parentID,attr"`
		// Restricted is whether or not the object can be modified remotely (e.g. by a Control Point).
		Restricted string `xml:"restricted,attr"`
		ChildCount string `xml:"childCount,attr"`
		Searchable string `xml:"searchable,attr"`

		Title       string `xml:"http://purl.org/dc/elements/1.1/ title"`
		Class       Class  `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ class"`
		StorageUsed string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ storageUsed"`
	}
	Item struct {
		ID       Object `xml:"id,attr"`
		ParentID Object `xml:"parentID,attr"`
		RefID    string `xml:"refID,attr"`
		// Restricted is whether or not the object can be modified remotely (e.g. by a Control Point).
		Restricted string `xml:"restricted,attr"`

		Title   string `xml:"http://purl.org/dc/elements/1.1/ title,omitempty"`
		Creator string `xml:"http://purl.org/dc/elements/1.1/ creator,omitempty"`
		Date    string `xml:"http://purl.org/dc/elements/1.1/ date,omitempty"`
		Class   Class  `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ class,omitempty"`
		// Artist is the name of an artist.
		// The <artist> tag also has a role= attribute.
		Artist string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ artist,omitempty"`
		// Actor is the name of an actor.
		// The <actor> tag also has a role= attribute.
		Actor string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ actor,omitempty"`
		// Author is the name of an author.
		// The <author> tag also has a role= attribute.
		Author              string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ author,omitempty"`
		Director            string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ director,omitempty"`
		Genre               string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ genre,omitempty"`
		Album               string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ album,omitempty"`
		Playlist            string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ playlist,omitempty"`
		OriginalTrackNumber uint   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ originalTrackNumber,omitempty"`

		AlbumArtURI []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ albumArtURI,omitempty"`
		LyricsURI   string   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ lyricsURI,omitempty"`

		Description     string `xml:"http://purl.org/dc/elements/1.1/ description,omitempty"`
		LongDescription string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ description,omitempty"`

		IconURI string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ icon,omitempty"`

		UserAnnotations []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ userAnnotation,omitempty"`

		Resources []Resource `xml:"res,omitempty"`
	}
	Resource struct {
		URI          string        `xml:",innerxml"`
		ProtocolInfo *ProtocolInfo `xml:"protocolInfo,attr"`
		Size         uint          `xml:"size,attr"`

		// Duration is of the form H+:MM:SS[.F+] or H+:MM:SS[.F0/F1], where:
		// H+ is 0 or more digits for hours,
		// MM is exactly 2 digits for minutes,
		// SS is exactly 2 digits for seconds,
		// F+ is 0 or more digits for fractional seconds,
		// F0/F1 is a fraction, F0 & F1 are at least 1 digit, and F0/F1 < 1.
		Duration string `xml:"duration,attr"`
		// Bitrate is in bits/second.
		Bitrate uint `xml:"bitrate,attr"`
		// SampleFrequency is in Hz.
		SampleFrequency uint `xml:"sampleFrequency,attr"`
		AudioChannels   uint `xml:"nrAudioChannels,attr"`
		// Resolution of the resource of the form [0-9]+x[0-9]+, e.g. 4x2.
		Resolution string `xml:"resolution,attr"`
	}
)

const (
	didlLiteSchema   = "urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/"
	dublinCoreSchema = "http://purl.org/dc/elements/1.1/"
)

func ParseDIDL(raw []byte) (*DIDL, error) {
	got := &DIDL{}
	if err := xml.Unmarshal(raw, got); err != nil {
		return nil, fmt.Errorf("could not unmarshal raw XML: %v", err)
	}
	return got, nil
}

// DIDLForURI returns a minimal DIDL sufficient to get media to play with just a URI.
//
// NB: It may not be enough, e.g. my TV needs more information about the video
// codec than can be inferred from just the URI.
func DIDLForURI(uri string) (*DIDL, error) {
	protocolInfo, err := ProtocolInfoForURI(uri)
	if err != nil {
		return nil, fmt.Errorf("could not create ProtocolInfo: %w", err)
	}
	class, err := ClassForURI(uri)
	if err != nil {
		return nil, fmt.Errorf("could not find item class: %w", err)
	}
	return &DIDL{
		Items: []Item{{
			Title: uri,
			Class: class,
			Resources: []Resource{{
				ProtocolInfo: protocolInfo,
				URI:          uri,
			}},
		}},
	}, nil
}

// URIForProtocolInfos finds a URI from an item that matches a set of valid ProtocolInfos.
// TODO: Return the "best" supported URI instead of just the first.
func (item *Item) URIForProtocolInfos(infos []*ProtocolInfo) (string, bool) {
	for _, resource := range item.Resources {
		resInfo := resource.ProtocolInfo
		for _, info := range infos {
			if resInfo.Protocol == info.Protocol &&
				resInfo.Network == info.Network &&
				resInfo.ContentFormat == info.ContentFormat {
				return resource.URI, true
			}
		}
	}
	return "", false
}
