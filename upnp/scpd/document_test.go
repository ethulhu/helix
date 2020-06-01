// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package scpd

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		scpd Document
		want string
	}{
		{
			scpd: Document{
				SpecVersion: SpecVersion{
					Major: 2,
					Minor: 1,
				},
				StateVariables: []StateVariable{
					{
						Name:                "TransportState",
						SendEventsAttribute: No,
						DataType:            "string",
						AllowedValues: &AllowedValues{
							Values: []string{
								"STOPPED",
								"PLAYING",
							},
						},
					},
					{
						Name:                "NumberOfTracks",
						SendEventsAttribute: Yes,
						DataType:            "ui4",
						AllowedValueRange: &AllowedValueRange{
							Minimum: 0,
							Maximum: 100,
							Step:    1,
						},
					},
					{
						Name:                "TrackCount",
						SendEventsAttribute: Yes,
						DataType:            "ui4",
						AllowedValueRange: &AllowedValueRange{
							Minimum: 0,
						},
					},
				},
				Actions: []Action{
					{
						Name: "SetAVTransportURI",
						Arguments: []Argument{
							{
								Name:                 "InstanceID",
								Direction:            In,
								RelatedStateVariable: "A_ARG_TYPE_InstanceID",
							},
							{
								Name:                 "CurrentURI",
								Direction:            Out,
								RelatedStateVariable: "A_ARG_TYPE_CurrentURI",
							},
						},
					},
				},
			},
			want: `<scpd xmlns="urn:schemas-upnp-org:service-1-0">
  <specVersion>
    <major>2</major>
    <minor>1</minor>
  </specVersion>
  <serviceStateTable>
    <stateVariable>
      <name>TransportState</name>
      <sendEventsAttribute>no</sendEventsAttribute>
      <dataType>string</dataType>
      <allowedValueList>
        <allowedValues>STOPPED</allowedValues>
        <allowedValues>PLAYING</allowedValues>
      </allowedValueList>
    </stateVariable>
    <stateVariable>
      <name>NumberOfTracks</name>
      <sendEventsAttribute>yes</sendEventsAttribute>
      <dataType>ui4</dataType>
      <allowedValueRange>
        <minimum>0</minimum>
        <maximum>100</maximum>
        <step>1</step>
      </allowedValueRange>
    </stateVariable>
    <stateVariable>
      <name>TrackCount</name>
      <sendEventsAttribute>yes</sendEventsAttribute>
      <dataType>ui4</dataType>
      <allowedValueRange>
        <minimum>0</minimum>
      </allowedValueRange>
    </stateVariable>
  </serviceStateTable>
  <actionList>
    <action>
      <name>SetAVTransportURI</name>
      <argumentList>
        <argument>
          <name>InstanceID</name>
          <direction>in</direction>
          <relatedStateVariable>A_ARG_TYPE_InstanceID</relatedStateVariable>
        </argument>
        <argument>
          <name>CurrentURI</name>
          <direction>out</direction>
          <relatedStateVariable>A_ARG_TYPE_CurrentURI</relatedStateVariable>
        </argument>
      </argumentList>
    </action>
  </actionList>
</scpd>`,
		},
	}

	for i, tt := range tests {
		bytes, err := xml.MarshalIndent(tt.scpd, "", "  ")
		if err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}
		got := string(bytes)

		if got != tt.want {
			t.Errorf("[%d]: got:\n\n%v\n\nwanted:\n\n%v", i, got, tt.want)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		raw  string
		want Document
	}{
		{
			raw: `<scpd xmlns="urn:schemas-upnp-org:service-1-0">
  <specVersion>
    <major>2</major>
    <minor>1</minor>
  </specVersion>
  <serviceStateTable>
    <stateVariable>
      <name>TransportState</name>
      <sendEventsAttribute>no</sendEventsAttribute>
      <dataType>string</dataType>
      <allowedValueList>
        <allowedValues>STOPPED</allowedValues>
        <allowedValues>PLAYING</allowedValues>
      </allowedValueList>
    </stateVariable>
    <stateVariable>
      <name>NumberOfTracks</name>
      <sendEventsAttribute>yes</sendEventsAttribute>
      <dataType>ui4</dataType>
      <allowedValueRange>
        <minimum>0</minimum>
        <maximum>100</maximum>
        <step>1</step>
      </allowedValueRange>
    </stateVariable>
  </serviceStateTable>
  <actionList>
    <action>
      <name>SetAVTransportURI</name>
      <argumentList>
        <argument>
          <name>InstanceID</name>
          <direction>in</direction>
          <relatedStateVariable>A_ARG_TYPE_InstanceID</relatedStateVariable>
        </argument>
        <argument>
          <name>CurrentURI</name>
          <direction>out</direction>
          <relatedStateVariable>A_ARG_TYPE_CurrentURI</relatedStateVariable>
        </argument>
      </argumentList>
    </action>
  </actionList>
</scpd>`,
			want: Document{
				XMLName: xml.Name{
					Local: "scpd",
					Space: xmlns,
				},
				SpecVersion: SpecVersion{
					Major: 2,
					Minor: 1,
				},
				StateVariables: []StateVariable{
					{
						Name:                "TransportState",
						SendEventsAttribute: No,
						DataType:            "string",
						AllowedValues: &AllowedValues{
							Values: []string{
								"STOPPED",
								"PLAYING",
							},
						},
					},
					{
						Name:                "NumberOfTracks",
						SendEventsAttribute: Yes,
						DataType:            "ui4",
						AllowedValueRange: &AllowedValueRange{
							Minimum: 0,
							Maximum: 100,
							Step:    1,
						},
					},
				},
				Actions: []Action{
					{
						Name: "SetAVTransportURI",
						Arguments: []Argument{
							{
								Name:                 "InstanceID",
								Direction:            In,
								RelatedStateVariable: "A_ARG_TYPE_InstanceID",
							},
							{
								Name:                 "CurrentURI",
								Direction:            Out,
								RelatedStateVariable: "A_ARG_TYPE_CurrentURI",
							},
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		got := Document{}
		if err := xml.Unmarshal([]byte(tt.raw), &got); err != nil {
			t.Fatalf("[%d]: got error: %v", i, err)
		}

		if !reflect.DeepEqual(got, tt.want) {
			gotBytes, _ := xml.MarshalIndent(got, "", "  ")
			wantBytes, _ := xml.MarshalIndent(tt.want, "", "  ")

			t.Errorf("[%d]: got:\n\n%s\n\nwanted:\n\n%s", i, gotBytes, wantBytes)
		}
	}
}
