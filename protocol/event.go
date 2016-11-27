package protocol

import (
	"encoding/xml"
	"io"
)

// BasicEvent is the xml payload sent to /upnp/control/basicevent1
type BasicEvent struct {
	XMLName struct{} `xml:"Envelope"`
	Body    BasicEventBody
}

type BasicEventBody struct {
	SetBinaryState BasicEventSetBinaryState
}

type BasicEventSetBinaryState struct {
	BinaryState *int
}

func DecodeBasicEvent(reader io.Reader) (event *BasicEvent, err error) {
	event = &BasicEvent{}
	err = xml.NewDecoder(reader).Decode(event)
	return event, err
}

func (be *BasicEvent) BinaryState() *int {
	return be.Body.SetBinaryState.BinaryState
}
