package embednats

import (
	"errors"
	"fmt"
	"github.com/nats-io/gnatsd/auth"
	"github.com/nats-io/gnatsd/logger"
	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"
	"math/rand"
	"sync"
	"time"
)

type Server struct {
	s    *server.Server
	auth *auth.Plain
	sWg  *sync.WaitGroup
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func NewServer(serverIP string) *Server {

	emb := &Server{}
	emb.sWg = &sync.WaitGroup{}

	emb.auth = &auth.Plain{Username: "natu", Password: RandStringRunes(50)}

	natsopts := &server.Options{Port: -1, Host: serverIP}

	emb.s = server.New(natsopts)
	emb.s.SetClientAuthMethod(emb.auth)
	if false {
		// TODO : revisit if logging is needed
		natslogger := logger.NewStdLogger(true, false, false, false, true)
		emb.s.SetLogger(natslogger, false, false)
	}
	emb.sWg.Add(1)
	go func() {
		emb.s.Start()
		emb.sWg.Done()
	}()

	// wait for server to start to get the dynamic port
	start := time.Now()
	for emb.s.GetListenEndpoint() == "" {
		time.Sleep(10 * time.Millisecond)
		if time.Since(start) > (20 * time.Second) {
			panic(errors.New("nats server failed to return endpoint in 20 seconds"))
		}
	}

	return emb
}

func (s *Server) GetEndpoint() string {
	return fmt.Sprintf("nats://%v:%v@%v", s.auth.Username, s.auth.Password, s.s.GetListenEndpoint())
}

func (s *Server) Connect() (*nats.Conn, error) {
	opts := nats.DefaultOptions
	opts.Servers = []string{s.GetEndpoint()}
	return opts.Connect()
}

func (s *Server) Publish(subject string, data []byte) error {
	ns, err := s.Connect()
	if err != nil {
		return err
	}
	defer ns.Close()
	defer ns.Flush()

	err = ns.Publish(subject, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown() {
	s.s.Shutdown()
	s.sWg.Wait()
}
