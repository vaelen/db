package main

import (
	"fmt"
	"os"
	"path/filepath"
//	"time"
	"github.com/vaelen/db/server"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't get working directory: %s\n", err.Error())
		os.Exit(1)
	}

	dbPath := filepath.Join(wd, "db")
	err = os.MkdirAll(dbPath, 0770)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't create data directory: %s\n", err.Error())
		os.Exit(2)
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't stat data directory: %s\n", err.Error())
		os.Exit(3)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Data directory is not a directory: %s\n", dbPath)
		os.Exit(4)
	}
	
	s := server.New(os.Stderr, dbPath)
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
