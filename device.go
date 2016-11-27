package vwego

import (
	"fmt"
	"github.com/mlctrez/vwego/protocol"
	"github.com/nats-io/nats"
	"log"
	"net"
	"net/http"
	"runtime"
)

type Device struct {
	vServer  *VwegoServer
	natsConn *nats.EncodedConn
	UUID     string
	UU       string
	Name     string
	Port     int
}

type DeviceEvent struct {
	UUID    string
	Name    string
	Command string
}

type LogData map[string]interface{}

func (d *Device) StartServer(vServer *VwegoServer) {

	d.vServer = vServer

	d.connectNats()

	server := &http.Server{Addr: fmt.Sprintf("%s:%d", d.vServer.ServerIP, d.Port), Handler: d}

	d.Debug(&LogData{"msg": "starting server", "addr": server.Addr, "name": d.Name})

	server.ListenAndServe()
}

func (d *Device) DiscoveryRequest(m *protocol.DiscoveryRequest) {
	addr, err := net.ResolveUDPAddr("udp4", m.RemoteHost)
	if err != nil {
		d.Error(&LogData{"method": "ResolveUDPAddr", "err": err.Error()})
	} else {
		d.SendDiscoveryResponse(addr)
	}
}

func (d *Device) connectNats() {
	c, err := d.vServer.EncodedConnection()
	if err != nil {
		log.Println("error connecting to nats server", err)
		return
	}
	d.natsConn = c
	d.natsConn.Subscribe("protocol.DiscoveryRequest", d.DiscoveryRequest)
}

func (d *Device) ServeSettingsXml(rw http.ResponseWriter, req *http.Request) {
	settings := protocol.NewSettings(d.Name, d.UU)
	d.natsConn.Publish("protocol.Settings", settings)
	err := settings.Write(rw)
	if err != nil {
		d.Error(&LogData{"method": "settings.Write", "err": err.Error()})
	}
}

func (d *Device) PublishEvent(command string) {
	e := &DeviceEvent{Name: d.Name, UUID: d.UUID, Command: command}
	d.natsConn.Publish("device.event", e)
}

func (d *Device) HandleBasicEvent(rw http.ResponseWriter, req *http.Request) {
	event, err := protocol.DecodeBasicEvent(req.Body)
	if event.BinaryState() != nil {
		switch *event.BinaryState() {
		case 1:
			d.PublishEvent("ON")
		case 0:
			d.PublishEvent("OFF")
		}
	} else {
		d.Error(&LogData{"method": "protocol.DecodeBasicEvent", "err": err.Error()})
	}
}

func (d *Device) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	reqURL := req.URL.String()

	d.Debug(&LogData{"DeviceUUID": d.UUID, "DeviceName": d.Name, "RemoteAddr": req.RemoteAddr, "URL": reqURL})

	switch reqURL {
	case "/settings.xml":
		d.ServeSettingsXml(rw, req)
	case "/upnp/control/basicevent1":
		d.HandleBasicEvent(rw, req)
	default:
		d.Debug(&LogData{"msg": "unknown request url", "URL": reqURL})
	}
}

func (d *Device) SendDiscoveryResponse(remoteAddress *net.UDPAddr) {
	con, err := net.DialUDP("udp4", nil, remoteAddress)

	if err != nil {
		d.Error(&LogData{"method": "net.DialUDP", "err": err.Error(), "raddr": remoteAddress})
		return
	}
	defer con.Close()

	parms := &protocol.DiscoveryResponseParams{
		DeviceName: d.Name,
		ServerIP:   d.vServer.ServerIP,
		ServerPort: d.Port,
		UUID:       d.UUID,
		UU:         d.UU,
	}

	drBytes := protocol.BuildDiscoveryResponse(parms)
	d.Debug(&LogData{"DiscoveryResponseParams": parms})
	con.Write(drBytes)
}

func (d *Device) Error(data *LogData) {
	msg := map[string]interface{}{"context": getContext(), "data": data}
	d.natsConn.Publish("error.device", msg)
}

func (d *Device) Debug(data *LogData) {
	msg := map[string]interface{}{"context": getContext(), "data": data}
	d.natsConn.Publish("debug.device", msg)
}

func getContext() string {
	pc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	context := "unknown"
	if ok && details != nil {
		context = details.Name()
	}
	return context
}
