package vwego

import (
	"encoding/json"
	"github.com/mlctrez/vwego/embednats"
	"github.com/mlctrez/vwego/protocol"
	"github.com/nats-io/nats"
	"github.com/satori/go.uuid"
	"log"
	"net"
	"os"
	"strings"
)

var uPnPAddress = "239.255.255.250:1900"

// VwegoServer is the virtual we go server struct
type VwegoServer struct {
	ServerIP   string
	ConfigPath string
	NatsServer *embednats.Server
	Config     *DeviceConfig
}

// DeviceConfig is the format of the configuration file
type DeviceConfig struct {
	Devices []*Device
}

func listenUPnP() (conn *net.UDPConn, err error) {
	addr, err := net.ResolveUDPAddr("udp4", uPnPAddress)
	if err != nil {
		return nil, err
	}
	conn, err = net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (s *VwegoServer) createDiscoveryListener() (listener func(), err error) {

	conn, err := listenUPnP()
	if err != nil {
		return nil, err
	}

	listener = func() {
		var buf [1024]byte
		for {
			packetLength, remote, err := conn.ReadFromUDP(buf[:])
			if err != nil {
				continue
			}
			dr := protocol.ParseDiscoveryRequest(buf[:packetLength])

			if dr.IsDeviceRequest() {
				dr.RemoteHost = remote.String()
				dj, err := json.Marshal(dr)
				if err != nil {
					continue
				}
				s.NatsServer.Publish("protocol.DiscoveryRequest", dj)
			}
		}
	}
	return listener, nil
}

func (s *VwegoServer) logMessages() {
	nc, err := s.NatsServer.Connect()
	if err != nil {
		log.Println("unable to connect")
		return
	}

	cb := make(chan *nats.Msg, 10)
	nc.ChanSubscribe(">", cb)

	for !nc.IsClosed() {
		select {
		case m := <-cb:
			log.Println(m.Subject, string(m.Data))
		}
	}
}

func (s *VwegoServer) Run() {
	log.SetOutput(os.Stdout)

	s.NatsServer = embednats.NewServer(s.ServerIP)

	go s.logMessages()

	err := s.ReadConfig()
	if err != nil {
		panic(err)
	}

	for _, device := range s.Config.Devices {
		go device.StartServer(s)
	}

	listener, err := s.createDiscoveryListener()
	if err != nil {
		panic(err)
	}
	listener()
}

func (s *VwegoServer) EncodedConnection() (ec *nats.EncodedConn, err error) {
	nc, err := s.NatsServer.Connect()
	if err != nil {
		return nil, err
	}
	ec, err = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}
	return ec, nil
}

func (s *VwegoServer) ReadConfig() error {
	f, err := os.Open(s.ConfigPath)
	if err != nil {
		return err
	}
	srv := &DeviceConfig{}
	err = json.NewDecoder(f).Decode(srv)
	if err != nil {
		return err
	}
	s.Config = srv
	return nil
}

func CreateConfig(devices []string, path string) error {
	d := &DeviceConfig{Devices: make([]*Device, 0)}
	for idx, name := range devices {
		u := uuid.NewV4()
		uParts := strings.Split(u.String(), "-")
		uu := uParts[len(uParts)-1]

		d.Devices = append(d.Devices, &Device{
			Name: name,
			UUID: u.String(),
			UU:   uu,
			Port: idx + 11000,
		})
	}
	return d.SaveConfig(path)
}

func (srv *DeviceConfig) SaveConfig(path string) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	bo, err := json.MarshalIndent(srv, "", "  ")
	if err != nil {
		return err
	}

	f.Write(bo)
	return nil
}
