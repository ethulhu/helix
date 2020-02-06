package avtransport

import (
	"encoding/xml"
)

type (
	playRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Play"`
		InstanceID int      `xml:"InstanceID"`
		Speed      string   `xml:"Speed"`
	}
	pauseRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Pause"`
		InstanceID int      `xml:"InstanceID"`
	}
	nextRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Next"`
		InstanceID int      `xml:"InstanceID"`
	}
	previousRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Previous"`
		InstanceID int      `xml:"InstanceID"`
	}
	stopRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Stop"`
		InstanceID int      `xml:"InstanceID"`
	}
	seekRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Seek"`
		InstanceID int      `xml:"InstanceID"`
		Unit       SeekMode `xml:"Unit"`
		Target     string   `xml:"Target"`
	}

	setAVTransportURIRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetAVTransportURI"`
		InstanceID int      `xml:"InstanceID"`
		CurrentURI string   `xml:"CurrentURI"`
		// CurrentMetadata is a DIDL-Lite document.
		CurrentMetadata []byte `xml:"CurrentURIMetaData"`
	}
	setNextAVTransportURIRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetNextAVTransportURI"`
		InstanceID int      `xml:"InstanceID"`
		NextURI    string   `xml:"NextURI"`
		// NextMetadata is a DIDL-Lite document.
		NextMetadata []byte `xml:"NextURIMetaData"`
	}

	getMediaInfoRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetMediaInfo"`
		InstanceID int      `xml:"InstanceID"`
	}
	getMediaInfoResponse struct {
		NumTracks  int    `xml:"NrTracks"`
		Duration   string `xml:"MediaDuration"`
		CurrentURI string `xml:"CurrentURI"`
		// CurrentMetadata is a DIDL-Lite document.
		CurrentMetadata []byte `xml:"CurrentURIMetaData"`
		NextURI         string `xml:"NextURI"`
		// NextMetadata is a DIDL-Lite document.
		NextMetadata []byte `xml:"NextURIMetaData"`
	}

	getPositionInfoRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetPositionInfo"`
		InstanceID int      `xml:"InstanceID"`
	}
	getPositionInfoResponse struct {
		CurrentTrack  string `xml:"Track"`
		Duration      string `xml:"TrackDuration"`
		Metadata      []byte `xml:"TrackMetaData"`
		URI           string `xml:"TrackURI"`
		RelativeTime  string `xml:"RelTime"`
		AbsoluteTime  string `xml:"AbsTime"`
		RelativeCount string `xml:"RelCount"`
		AbsoluteCount string `xml:"AbsCount"`
	}

	getTransportInfoRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetTransportInfo"`
		InstanceID int      `xml:"InstanceID"`
	}
	getTransportInfoResponse struct {
		TransportState  State  `xml:"CurrentTransportState"`
		TransportStatus string `xml:"CurrentTransportStatus"`
		Speed           string `xml:"CurrentSpeed"`
	}
)

const (
	setAVTransportURI     = "SetAVTransportURI"
	setNextAVTransportURI = "SetNextAVTransportURI"

	getMediaInfo          = "GetMediaInfo"
	getTransportInfo      = "GetTransportInfo"
	getPositionInfo       = "GetPositionInfo"
	getDeviceCapabilities = "GetDeviceCapabilities"
	getTransportSettings  = "GetTransportSettings"

	play     = "Play"
	pause    = "Pause"
	next     = "Next"
	previous = "Previous"
	record   = "Record"
	seek     = "Seek"
	stop     = "Stop"

	setPlayMode                = "SetPlayMode"
	setRecordQualityMode       = "SetRecordQualityMode"
	getCurrentTransportActions = "getCurrentTransportActions"
)
