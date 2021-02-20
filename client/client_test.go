package main

import (
	cfg "go-proxy/config"
	"log"
	"testing"
)

func TestClient(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cfg.Init("../bin/conf.yml", &conf)

	log.Printf("The client has started, ready to connect to %s\n", conf.ServerAddr)
	cli := NewRPClient(conf.ServerAddr, conf.MaxConn, conf.Token)
	_ = cli.Start()
}
