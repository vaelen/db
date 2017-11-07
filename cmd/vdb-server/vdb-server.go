/******
This file is part of Vaelen/DB.

Copyright 2017, Andrew Young <andrew@vaelen.org>

    Vaelen/DB is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

    Vaelen/DB is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
along with Vaelen/DB.  If not, see <http://www.gnu.org/licenses/>.
******/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/vaelen/db/server"
	"google.golang.org/grpc"
	"net"
	"log"
	"os/signal"
	"github.com/vaelen/db/api"
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

	lis, err := net.Listen("tcp", ":5555")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterDatabaseServer(grpcServer, s)

	// Handle signals nicely
	signalHandler := make(chan os.Signal)
	signal.Notify(signalHandler, os.Interrupt, os.Kill)
	go func(s *server.DBServer) {
		for {
			select {
			case sig := <-signalHandler:
				// Handle signal
				s.Logger.Printf("Caught Signal: %s\n", sig)
				switch sig {
				case os.Interrupt:
					s.Stop()
					grpcServer.Stop()
					break
				case os.Kill:
					s.Stop()
					grpcServer.Stop()
					break
				default:
					s.Logger.Printf("Signal Ignored\n")
				}
			}
		}
	}(s)
	grpc.WithInsecure()
	grpcServer.Serve(lis)
}
