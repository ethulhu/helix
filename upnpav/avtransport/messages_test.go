package avtransport

import (
	"testing"

	"github.com/ethulhu/helix/upnp/scpd"
)

func TestSCPD(t *testing.T) {
	actions := []struct {
		name     string
		req, rsp interface{}
	}{
		{play, playRequest{}, playResponse{}},
		{stop, stopRequest{}, stopResponse{}},
		{next, nextRequest{}, nextResponse{}},
		{stop, stopRequest{}, stopResponse{}},
		{previous, previousRequest{}, previousResponse{}},
		{record, recordRequest{}, recordResponse{}},
		{seek, seekRequest{}, seekResponse{}},
		{getCurrentTransportActions, getCurrentTransportActionsRequest{}, getCurrentTransportActionsResponse{}},
		{setPlayMode, setPlayModeRequest{}, setPlayModeResponse{}},
		{setRecordQualityMode, setRecordQualityModeRequest{}, setRecordQualityModeResponse{}},
		{setAVTransportURI, setAVTransportURIRequest{}, setAVTransportURIResponse{}},
		{setNextAVTransportURI, setNextAVTransportURIRequest{}, setNextAVTransportURIResponse{}},
		{getDeviceCapabilities, getDeviceCapabilitiesRequest{}, getDeviceCapabilitiesResponse{}},
		{getMediaInfo, getMediaInfoRequest{}, getMediaInfoResponse{}},
		{getPositionInfo, getPositionInfoRequest{}, getPositionInfoResponse{}},
		{getTransportInfo, getTransportInfoRequest{}, getTransportInfoResponse{}},
		{getTransportSettings, getTransportSettingsRequest{}, getTransportSettingsResponse{}},
	}

	var docs []scpd.Document
	for _, action := range actions {
		doc, err := scpd.FromAction(action.name, action.req, action.rsp)
		if err != nil {
			t.Errorf("SCPD definition for action %q is broken: %v", action.name, err)
		}
		docs = append(docs, doc)
	}

	if _, err := scpd.Merge(docs...); err != nil {
		t.Errorf("could not merge SCPDs: %v", err)
	}
}
