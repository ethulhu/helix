// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package avtransport

import (
	"encoding/xml"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/xmltypes"
)

type (
	setAVTransportURIRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetAVTransportURI"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
		CurrentURI string   `xml:"CurrentURI" scpd:"AVTransportURI,string"`

		// CurrentMetadata is a DIDL-Lite document.
		CurrentMetadata []byte `xml:"CurrentURIMetaData" scpd:"AVTransportURIMetaData,string"`
	}
	setAVTransportURIResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetAVTransportURIResponse"`
	}

	setNextAVTransportURIRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetNextAVTransportURI"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
		NextURI    string   `xml:"NextURI"    scpd:"NextAVTransportURI,string"`

		// NextMetadata is a DIDL-Lite document.
		NextMetadata []byte `xml:"NextURIMetaData" scpd:"NextAVTransportURIMetaData,string"`
	}
	setNextAVTransportURIResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetNextAVTransportURIResponse"`
	}

	getMediaInfoRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetMediaInfo"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	getMediaInfoResponse struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetMediaInfoResponse"`
		TrackCount uint     `xml:"NrTracks"      scpd:"NumberOfTracks,ui4,min=4"`
		Duration   string   `xml:"MediaDuration" scpd:"CurrentMediaDuration,string"`
		CurrentURI string   `xml:"CurrentURI"    scpd:"AVTransportURI,string"`
		NextURI    string   `xml:"NextURI"       scpd:"NextAVTransportURI,string"`

		// Metadata is a DIDL-Lite document.
		CurrentMetadata []byte `xml:"CurrentURIMetaData" scpd:"AVTransportURIMetaData,string"`
		NextMetadata    []byte `xml:"NextURIMetaData"    scpd:"NextAVTransportURIMetaData,string"`

		PlayMedium   string `xml:"PlayMedium"   scpd:"PlaybackStorageMedium,string"`
		RecordMedium string `xml:"RecordMedium" scpd:"RecordStorageMedium,string"`
		WriteStatus  string `xml:"WriteStatus"  scpd:"RecordMediumWriteStatus,string,WRITEABLE|PROTECTED|NOT_WRITEABLE|UNKNOWN"`
	}

	getTransportInfoRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetTransportInfo"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	getTransportInfoResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetTransportInfoResponse"`
		State   State    `xml:"CurrentTransportState"  scpd:"TransportState,string,STOPPED|PLAYING|TRANSITIONING|PAUSED_PLAYBACK|PAUSED_RECORDING|RECORDING|NO_MEDIA_PRESENT"`
		Status  Status   `xml:"CurrentTransportStatus" scpd:"TransportStatus,string,OK|ERROR_OCCURRED"`
		Speed   string   `xml:"CurrentSpeed"           scpd:"TransportPlaySpeed,string,1"`
	}

	getPositionInfoRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetPositionInfo"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	getPositionInfoResponse struct {
		XMLName       xml.Name        `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetPositionInfoResponse"`
		CurrentTrack  uint            `xml:"Track"         scpd:"CurrentTrack,ui4,min=0,step=1"`
		Duration      upnpav.Duration `xml:"TrackDuration" scpd:"CurrentTrackDuration,string"`
		Metadata      []byte          `xml:"TrackMetaData" scpd:"CurrentTrackMetaData,string"`
		URI           string          `xml:"TrackURI"      scpd:"CurrentTrackURI,string"`
		RelativeTime  upnpav.Duration `xml:"RelTime"       scpd:"RelativeTimePosition,string"`
		AbsoluteTime  upnpav.Duration `xml:"AbsTime"       scpd:"AbsoluteTimePosition,string"`
		RelativeCount int             `xml:"RelCount"      scpd:"RelativeCounterPosition,i4"`
		AbsoluteCount int             `xml:"AbsCount"      scpd:"AbsoluteCounterPosition,i4"`
	}

	getDeviceCapabilitiesRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetDeviceCapabilities"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	getDeviceCapabilitiesResponse struct {
		XMLName            xml.Name                       `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetDeviceCapabilitiesResponse"`
		PlayMedia          xmltypes.CommaSeparatedStrings `xml:"PlayMedia"       scpd:"PossiblePlaybackStorageMedia,string"`
		RecordMedia        xmltypes.CommaSeparatedStrings `xml:"RecMedia"        scpd:"PossibleRecordStorageMedia,string"`
		RecordQualityModes xmltypes.CommaSeparatedStrings `xml:"RecQualityModes" scpd:"PossibleRecordQualityModes,string"`
	}

	getTransportSettingsRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetTransportSettings"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	getTransportSettingsResponse struct {
		XMLName           xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetTransportSettingsResponse"`
		PlayMode          string   `xml:"PlayMode"       scpd:"CurrentPlayMode,string,NORMAL|SHUFFLE|REPEAT_ONE|REPEAT_ALL|RANDOM|DIRECT_1|INTRO"`
		RecordQualityMode string   `xml:"RecQualityMode" scpd:"CurrentRecordQualityMode,string,0:EP|1:LP|2:SP|0:BASIC|1:MEDIUM|2:HIGH"`
	}

	stopRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Stop"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	stopResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 StopResponse"`
	}

	playRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Play"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
		Speed      string   `xml:"Speed"      scpd:"TransportPlaySpeed,string,1"`
	}
	playResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 PlayResponse"`
	}

	pauseRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Pause"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	pauseResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 PauseResponse"`
	}

	recordRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Record"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	recordResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 RecordResponse"`
	}

	seekRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Seek"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
		Unit       SeekMode `xml:"Unit"       scpd:"A_ARG_TYPE_SeekMode,string,TRACK_NR|ABS_TIME|REL_TIME|ABS_COUNT|REL_COUNT|CHANNEL_FREQ|TAPE-INDEX|FRAME"`
		Target     string   `xml:"Target"     scpd:"A_ARG_TYPE_SeekTarget,string"`
	}
	seekResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SeekResponse"`
	}

	nextRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Next"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	nextResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 NextResponse"`
	}

	previousRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 Previous"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	previousResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 PreviousResponse"`
	}

	setPlayModeRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetPlayMode"`
		InstanceID int      `xml:"InstanceID"  scpd:"A_ARG_TYPE_InstanceID,ui4"`
		PlayMode   string   `xml:"NewPlayMode" scpd:"CurrentPlayMode,string,NORMAL|SHUFFLE|REPEAT_ONE|REPEAT_ALL|RANDOM|DIRECT_1|INTRO"`
	}
	setPlayModeResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetPlayModeResponse"`
	}

	setRecordQualityModeRequest struct {
		XMLName           xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetRecordQualityMode"`
		InstanceID        int      `xml:"InstanceID"           scpd:"A_ARG_TYPE_InstanceID,ui4"`
		RecordQualityMode string   `xml:"NewRecordQualityMode" scpd:"CurrentRecordQualityMode,string,0:EP|1:LP|2:SP|0:BASIC|1:MEDIUM|2:HIGH"`
	}
	setRecordQualityModeResponse struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 SetRecordQualityModeResponse"`
	}

	getCurrentTransportActionsRequest struct {
		XMLName    xml.Name `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetCurrentTransportActions"`
		InstanceID int      `xml:"InstanceID" scpd:"A_ARG_TYPE_InstanceID,ui4"`
	}
	getCurrentTransportActionsResponse struct {
		XMLName xml.Name                       `xml:"urn:schemas-upnp-org:service:AVTransport:1 GetCurrentTransportActionsResponse"`
		Actions xmltypes.CommaSeparatedStrings `xml:"Actions" scpd:"CurrentTransportActions,string"`
	}
)

const (
	setAVTransportURI     = "SetAVTransportURI"
	setNextAVTransportURI = "SetNextAVTransportURI"

	getDeviceCapabilities = "GetDeviceCapabilities"
	getMediaInfo          = "GetMediaInfo"
	getPositionInfo       = "GetPositionInfo"
	getTransportInfo      = "GetTransportInfo"
	getTransportSettings  = "GetTransportSettings"

	play     = "Play"
	pause    = "Pause"
	next     = "Next"
	previous = "Previous"
	record   = "Record"
	seek     = "Seek"
	stop     = "Stop"

	getCurrentTransportActions = "GetCurrentTransportActions"
	setPlayMode                = "SetPlayMode"
	setRecordQualityMode       = "SetRecordQualityMode"
)
