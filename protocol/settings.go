package protocol

import (
	"encoding/xml"
	"fmt"
	"io"
)

// Settings represents the response payload for /settings.xml
type Settings struct {
	XMLName struct{}       `xml:"root"`
	Device  SettingsDevice `xml:"device"`
}

type SettingsDevice struct {
	DeviceType   string `xml:"deviceType"`
	FriendlyName string `xml:"friendlyName"`
	Manufacturer string `xml:"manufacturer"`
	ModelName    string `xml:"modelName"`
	ModelNumber  string `xml:"modelNumber"`
	UDN          string `xml:"UDN"`
}

func NewSettings(friendlyName string, udn string) *Settings {
	s := &Settings{
		Device: SettingsDevice{
			DeviceType:   "urn:EmulatedSocket:device:controllee:1",
			FriendlyName: friendlyName,
			Manufacturer: "Belkin International Inc.",
			ModelName:    "Emulated Socket",
			ModelNumber:  "6.022140857",
			UDN:          fmt.Sprintf("uuid:Socket-1_0-%s", udn),
		},
	}
	return s
}

func (s *Settings) Write(writer io.Writer) error {
	_, err := writer.Write([]byte(xml.Header))
	if err != nil {
		return err
	}
	return xml.NewEncoder(writer).Encode(s)
}
