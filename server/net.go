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

package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/vaelen/db/storage"
	"github.com/vaelen/db/api"
)

// Server is an instance of the database server
type Server struct {
	Shutdown      chan bool
	SignalHandler chan os.Signal
	Logger        *log.Logger
	Storage       *storage.Storage
	listeners     []*listener
	logWriter     io.Writer
}

// ListenAddress is a network address that the server should listen on
type ListenAddress struct {
	NetworkType string
	Address     string
}

type listener struct {
	Shutdown chan bool
	l        net.Listener
}

// New creates a new instance of the database server
func New(logWriter io.Writer, dbPath string) *Server {
	return &Server{
		Shutdown:      make(chan bool),
		SignalHandler: make(chan os.Signal),
		Logger:        log.New(logWriter, "[NETWORK] ", log.LstdFlags),
		Storage:       storage.New(logWriter, dbPath),
		listeners:     make([]*listener, 0),
		logWriter:     logWriter,
	}
}

func newListener(l net.Listener) *listener {
	return &listener{
		Shutdown: make(chan bool),
		l:        l,
	}
}

// Start starts the database server
func (s *Server) Start(addresses []ListenAddress) {
	done := false
	s.Logger.Printf("Starting...\n")
	// Open network listeners
	for _, address := range addresses {
		l, err := s.listen(address)
		if err != nil {
			s.Logger.Printf("Error: %s\n", err.Error)
			continue
		}
		s.listeners = append(s.listeners, l)
	}

	// Handle signals nicely
	signal.Notify(s.SignalHandler, os.Interrupt, os.Kill)
	go func(s *Server) {
		for {
			select {
			case sig := <-s.SignalHandler:
				// Handle signal
				s.Logger.Printf("Caught Signal: %s\n", sig)
				switch sig {
				case os.Interrupt:
					s.Shutdown <- true
					break
				case os.Kill:
					s.Shutdown <- true
					break
				default:
					s.Logger.Printf("Signal Ignored\n")
				}
			}
		}
	}(s)

	s.Logger.Printf("Ready\n")

	// Main processing loop
	for {
		select {
		case <-s.Shutdown:
			// Stop the server
			s.Logger.Printf("Stopping...\n")
			done = true
			for _, l := range s.listeners {
				l.Shutdown <- true
			}
			if s.Storage != nil {
				s.Storage.Shutdown <- true
			}
		}
		if done {
			break
		}
	}
	s.Logger.Printf("Stopped\n")
}

func (s *Server) listen(address ListenAddress) (*listener, error) {
	l, err := net.Listen(address.NetworkType, address.Address)
	if err != nil {
		return nil, err
	}
	s.Logger.Printf("Listening on %s %s.\n", address.NetworkType, address.Address)
	listener := newListener(l)
	go s.acceptConnections(listener)
	return listener, nil
}

func (s *Server) acceptConnections(l *listener) {
	defer l.l.Close()
	for {
		select {
		case <-l.Shutdown:
			break
		default:
			tcpL, ok := l.l.(*net.TCPListener)
			if ok {
				tcpL.SetDeadline(time.Now().Add(1e9))
			}
			conn, err := l.l.Accept()
			if err != nil {
				//s.Logger.Printf("Error: %s\n", err.Error())
				continue
			}
			go s.connectionHandler(conn)
		}
	}
}

func (s *Server) connectionHandler(c net.Conn) {
	// We don't do much yet
	t := time.Now()

	m := api.Multiplexer{ Stream: c }

	for {
		command := api.Command{}
		response := api.Response{}

		// Read message
		buf, err := m.Receive()
		if err == io.EOF {
			// Done
			break
		}
		if err == nil {
			err = proto.Unmarshal(buf, &command)
		}

		if err != nil {
			response.Error = err.Error()
		} else {
			id := command.ID
			value := command.Value
			switch command.Type {
			case api.Command_TIME:
				response.Value = fmt.Sprintf("Hello! The time is currently %s.\n",
					t.Format(time.RFC3339))
			case api.Command_GET:
				v := s.get(id)
				response.ID = id
				response.Value = v
			case api.Command_SET:
				v := s.update(id, value)
				response.ID = id
				response.Value = v
			}
		}

		buf, err = proto.Marshal(&response)
		if err != nil {
			s.Logger.Printf("Error marshalling response: %s\n", err.Error())
		} else {
			err = m.Send(buf)
			if err != nil {
				s.Logger.Printf("Error sending response: %s\n", err.Error())
			}
		}
	}

	c.Close()
}

func (s *Server) get(id string) string {
	return s.Storage.Get(id)
}

func (s *Server) update(id string, value string) string {
	return s.Storage.Set(id, value)
}

func (s *Server) remove(id string) string {
	return s.Storage.Remove(id)
}
