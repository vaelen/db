package main

import (
	"os"
//	"time"
	"github.com/vaelen/db/server"
)

func main() {
	s := server.New(os.Stderr)
	addresses := make([]server.ListenAddress, 0, 1)
	addresses = append(addresses, server.ListenAddress { NetworkType: "tcp", Address: ":5555" })
/*
	go func(s *server.Server) {
		time.Sleep(time.Second * 10)
		s.Logger.Printf("Sending shutdown signal...\n")
		s.Shutdown <- true
	}(s)
*/
	s.Start(addresses)
}
