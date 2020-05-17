package upnpav
//go:generate go run ./internal/mk-marshal-didllite -out ./didllite.marshal.go
//go:generate go run ./internal/mk-unmarshal-didllite -out ./didllite.unmarshal.go

import (
	"fmt"
	"net/url"
	"time"

	"github.com/beevik/etree"
)

type (
	DIDLLite struct {
		Containers []Container `upnpav:"container"`
		Items      []Item      `upnpav:"item"`
	}

	ObjectID string

	Object struct {
		ID     ObjectID `upnpav:"id,attr"`
		Parent ObjectID `upnpav:"parentID,attr"`

		// Writeable == true is actually restricted == false, and vice versa.
		Writeable  bool `upnpav:"restricted,attr,inverse"`
		Searchable bool `upnpav:"searchable,attr"`

		Title string `upnpav:"dc:title"`
		Class Class  `upnpav:"upnp:class"`

		Description     string   `upnpav:"dc:description"`
		LongDescription string   `upnpav:"upnp:longDescription"`
		Icon            *url.URL `upnpav:"upnp:icon"`
		Region          string   `upnpav:"upnp:region"`
		AgeRating       string   `upnpav:"upnp:rating"`
		Rights          []string `upnpav:"dc:rights"`

		// Date is an ISO8601 date of the form yyyy-mm-dd.
		Date time.Time `upnpav:"dc:date"`

		// Language is an RFC1766 language, e.g. "en-US".
		Language string `upnpav:"dc:language"`

		// UserAnnotations is a "general-purpose tag where a user can annotate an object with some user-specific information".
		UserAnnotations []string `upnpav:"upnp:userAnnotation"`

		// TOC is an "identifier of an audio CD".
		TOC string `upnpav:"upnp:toc"`

		// WriteStatus can be one of: WRITEABLE, PROTECTED, NOT_WRITEABLE, UNKNOWN, MIXED.
		WriteStatus string `upnpav:"upnp:writeStatus"`
	}

	Container struct {
		Object

		ChildCount int `upnpav:"childCount,attr"`

		// Storage has a special value "-1" to represent "unknown".
		StorageTotalBytes        int    `upnpav:"upnp:storageTotal"`
		StorageUsedBytes         int    `upnpav:"upnp:storageUsed"`
		StorageFreeBytes         int    `upnpav:"upnp:storageFree"`
		StorageMaxPartitionBytes int    `upnpav:"upnp:storageMaxPartition"`
		StorageMedium            string `upnpav:"upnp:storageMedium"`
	}

	Item struct {
		Object

		// RefID is "ID property of the item being referred to".
		RefID string `upnpav:"refID,attr"`

		Creator string `upnpav:"dc:creator"`

		Artists      []Person `upnpav:"upnp:artist"`
		Actors       []Person `upnpav:"upnp:actor"`
		Authors      []Person `upnpav:"upnp:author"`
		Directors    []string `upnpav:"upnp:director"`
		Producers    []string `upnpav:"upnp:producer"`
		Publishers   []string `upnpav:"dc:publisher"`
		Contributors []string `upnpav:"dc:contributor"`

		// The following link to containers by the container title (e.g. object.container.playlist).
		Genres    []string `upnpav:"upnp:genre"`
		Albums    []string `upnpav:"upnp:album"`
		Playlists []string `upnpav:"upnp:playlist"`

		AlbumArtURI          []string `upnpav:"upnp:albumArtURI"`
		ArtistDiscographyURI string   `upnpav:"upnp:artistDiscographyURI"`
		LyricsURI            string   `upnpav:"upnp:lyricsURI"`
		RelationURI          string   `upnpav:"dc:relation"`

		TrackNumber int `upnpav:"upnp:originalTrackNumber"`

		Resources []Resource `upnpav:"res"`
	}
	Person struct {
		Name string `upnpav:",innerxml"`
		Role string `upnpav:"role,attr"`
	}

	Resource struct {
		URI          string        `upnpav:",innerxml"`
		ProtocolInfo *ProtocolInfo `upnpav:"protocolInfo,attr"`

		// Duration is of the form H+:MM:SS[.F+] or H+:MM:SS[.F0/F1], where:
		// H+ is 0 or more digits for hours,
		// MM is exactly 2 digits for minutes,
		// SS is exactly 2 digits for seconds,
		// F+ is 0 or more digits for fractional seconds,
		// F0/F1 is a fraction, F0 & F1 are at least 1 digit, and F0/F1 < 1.

		AudioChannels     uint          `upnpav:"nrAudioChannels,attr"`
		BitsPerSample     uint          `upnpav:"bitsPerSample,attr"`
		BitsPerSecond     uint          `upnpav:"bitrate,attr"`
		ColorDepth        uint          `upnpav:"colorDepth,attr"`
		Duration          time.Duration `upnpav:"duration,attr"`
		Resolution        *Resolution   `upnpav:"resolution,attr"`
		SampleFrequencyHz uint          `upnpav:"sampleFrequency,attr"`
		SizeBytes         uint          `upnpav:"size,attr"`

		// Protection is "some identification of a protection system used for the resource".
		Protection string `upnpav:"protection,attr"`
		// ImportURI is "URI via which the resource can be imported to the CDS via ImportResource() or HTTP POST".
		ImportURI string `upnpav:"importURI,attr"`
	}

	// Resolution of the resource of the form [0-9]+x[0-9]+, e.g. 4x2.
	Resolution struct {
		Height, Width int
	}
)

func ParseDIDLLite(src string) (*DIDLLite, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(src); err != nil {
		return nil, err
	}
	return unmarshalDIDLLite(doc)
}
func (d *DIDLLite) String() string {
	document := marshalDIDLLite(d)

	document.Indent(2)
	document.WriteSettings.CanonicalEndTags = true
	out, err := document.WriteToString()
	if err != nil {
		panic(fmt.Errorf("could not write to string: %v", err))
	}
	return out
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
func (item *Item) HasURI(uri string) bool {
	for _, resource := range item.Resources {
		if uri == resource.URI {
			return true
		}
	}
	return false
}
