package main

import (
	"flag"
	"github.com/mlctrez/vwego"
	"log"
)

func main() {

	serverIP := flag.String("serverIP", "", "the ip address to bind services")
	config := flag.String("config", "config.json", "path to the config.json file")

	flag.Parse()

	if *serverIP == "" {
		flag.Usage()
		log.Fatal("need to specify a serverIP parameter")
	}

	s := &vwego.WegoServer{ServerIP: *serverIP, ConfigPath: *config}
	s.Run()

}
