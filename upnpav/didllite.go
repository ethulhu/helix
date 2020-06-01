// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnpav

//go:generate go run ./internal/mk-marshal-didllite -out ./didllite.marshal.go
//go:generate go run ./internal/mk-unmarshal-didllite -out ./didllite.unmarshal.go

import (
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/ethulhu/helix/xmltypes"
)

type (
	didlliteWrapper struct {
		XMLName   xml.Name `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ DIDL-Lite"`
		XMLNSDC   string   `xml:"xmlns:dc,attr"`
		XMLNSUPnP string   `xml:"xmlns:upnp,attr"`
		XMLNSDLNA string   `xml:"xmlns:dlna,attr"`
		DIDLLite
	}
	DIDLLite struct {
		Containers []Container `xml:"container,omitempty"`
		Items      []Item      `xml:"item,omitempty"`
	}

	ObjectID string

	Object struct {
		ID     ObjectID `xml:"id,attr"`
		Parent ObjectID `xml:"parentID,attr"`

		// (Restricted == true) == (writeable == false).
		Restricted xmltypes.IntBool `xml:"restricted,attr"`
		Searchable xmltypes.IntBool `xml:"searchable,attr"`

		Title string `xml:"http://purl.org/dc/elements/1.1/ title,omitempty"`
		Class Class  `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ class,omitempty"`

		Description     string   `xml:"http://purl.org/dc/elements/1.1/ description,omitempty"`
		LongDescription string   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ longDescription,omitempty"`
		Icon            *url.URL `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ icon,omitempty"`
		Region          string   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ region,omitempty"`
		AgeRating       string   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ rating,omitempty"`
		Rights          []string `xml:"http://purl.org/dc/elements/1.1/ rights,omitempty"`
		Date            *Date    `xml:"http://purl.org/dc/elements/1.1/ date,omitempty"`

		// Language is an RFC1766 language, e.g. "en-US".
		Language string `xml:"http://purl.org/dc/elements/1.1/ language,omitempty"`

		// UserAnnotations is a "general-purpose tag where a user can annotate an object with some user-specific information".
		UserAnnotations []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ userAnnotation,omitempty"`

		// TOC is an "identifier of an audio CD".
		TOC string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ toc,omitempty"`

		// WriteStatus can be one of: WRITEABLE, PROTECTED, NOT_WRITEABLE, UNKNOWN, MIXED.
		WriteStatus string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ writeStatus,omitempty"`
	}

	Container struct {
		Object `xml:"foo"`

		ChildCount int `xml:"childCount,attr"`

		// Storage has a special value "-1" to represent "unknown".
		StorageTotalBytes        int    `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ storageTotal,omitempty"`
		StorageUsedBytes         int    `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ storageUsed,omitempty"`
		StorageFreeBytes         int    `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ storageFree,omitempty"`
		StorageMaxPartitionBytes int    `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ storageMaxPartition,omitempty"`
		StorageMedium            string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ storageMedium,omitempty"`
	}

	Item struct {
		Object

		// RefID is "ID property of the item being referred to".
		RefID string `xml:"refID,attr,omitempty"`

		Creator string `xml:"http://purl.org/dc/elements/1.1/ creator,omitempty"`

		Artists      []Person `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ artist,omitempty"`
		Actors       []Person `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ actor,omitempty"`
		Authors      []Person `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ author,omitempty"`
		Directors    []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ director,omitempty"`
		Producers    []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ producer,omitempty"`
		Publishers   []string `xml:"http://purl.org/dc/elements/1.1/ publisher,omitempty"`
		Contributors []string `xml:"http://purl.org/dc/elements/1.1/ contributor,omitempty"`

		// The following link to containers by the container title (e.g. object.container.playlist).
		Genres    []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ genre,omitempty"`
		Albums    []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ album",omitempty`
		Playlists []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ playlist,omitempty"`

		AlbumArtURI          []string `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ albumArtURI,omitempty"`
		ArtistDiscographyURI string   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ artistDiscographyURI,omitempty"`
		LyricsURI            string   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ lyricsURI,omitempty"`
		RelationURI          string   `xml:"http://purl.org/dc/elements/1.1/ relation,omitempty"`

		TrackNumber int `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ originalTrackNumber,omitempty"`

		Resources []Resource `xml:"res,omitempty"`
	}

	Person struct {
		Name string `xml:",innerxml"`
		Role string `xml:"role,attr,omitempty"`
	}

	Resource struct {
		URI          string        `xml:",innerxml"`
		ProtocolInfo *ProtocolInfo `xml:"protocolInfo,attr,omitempty"`

		AudioChannels     uint        `xml:"nrAudioChannels,attr,omitempty"`
		BitsPerSample     uint        `xml:"bitsPerSample,attr,omitempty"`
		BitsPerSecond     uint        `xml:"bitrate,attr,omitempty"`
		ColorDepth        uint        `xml:"colorDepth,attr,omitempty"`
		Duration          *Duration   `xml:"duration,attr,omitempty"`
		Resolution        *Resolution `xml:"resolution,attr,omitempty"`
		SampleFrequencyHz uint        `xml:"sampleFrequency,attr,omitempty"`
		SizeBytes         uint        `xml:"size,attr,omitempty"`

		// Protection is "some identification of a protection system used for the resource".
		Protection string `xml:"protection,attr,omitempty"`

		// ImportURI is "URI via which the resource can be imported to the CDS via ImportResource() or HTTP POST".
		ImportURI string `xml:"importURI,attr,omitempty"`
	}
)

func ParseDIDLLite(src string) (*DIDLLite, error) {
	doc := didlliteWrapper{}
	if err := xml.Unmarshal([]byte(src), &doc); err != nil {
		return nil, err
	}
	return &doc.DIDLLite, nil
}
func (d DIDLLite) String() string {
	doc := didlliteWrapper{
		XMLNSDC:   "http://purl.org/dc/elements/1.1/",
		XMLNSUPnP: "urn:schemas-upnp-org:metadata-1-0/upnp/",
		XMLNSDLNA: "urn:schemas-dlna-org:metadata-1-0/",
		DIDLLite:  d,
	}
	bytes, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("could not marshal DIDLLite: %v", err))
	}
	return xml.Header + string(bytes)
}

// DIDLForURI returns a minimal DIDL sufficient to get media to play with just a URI.
//
// NB: It may not be enough, e.g. my TV needs more information about the video
// codec than can be inferred from just the URI.
func DIDLForURI(uri string) (*DIDLLite, error) {
	protocolInfo, err := ProtocolInfoForURI(uri)
	if err != nil {
		return nil, fmt.Errorf("could not create ProtocolInfo: %w", err)
	}
	class, err := ClassForURI(uri)
	if err != nil {
		return nil, fmt.Errorf("could not find item class: %w", err)
	}
	return &DIDLLite{
		Items: []Item{{
			Object: Object{
				Title: uri,
				Class: class,
			},
			Resources: []Resource{{
				ProtocolInfo: protocolInfo,
				URI:          uri,
			}},
		}},
	}, nil
}

// URIForProtocolInfos finds a URI from an item that matches a set of valid ProtocolInfos.
// TODO: Return the "best" supported URI instead of just the first.
func (item Item) URIForProtocolInfos(infos []ProtocolInfo) (string, bool) {
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
func (item Item) HasURI(uri string) bool {
	for _, resource := range item.Resources {
		if uri == resource.URI {
			return true
		}
	}
	return false
}
