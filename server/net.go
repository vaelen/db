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
	encoder "encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/vaelen/db/storage"
)

type CommandType uint8

const (
	TimeCommand   CommandType = 1
	GetCommand    CommandType = 2
	UpdateCommand CommandType = 3
)

type Command struct {
	Type  CommandType
	ID    string
	Value string
}

type Response struct {
	ID    string
	Value string
	Error string
}

type Server struct {
	Shutdown      chan bool
	SignalHandler chan os.Signal
	Logger        *log.Logger
	Storage       *storage.Storage
	listeners     []*Listener
	logWriter     io.Writer
}

type ListenAddress struct {
	NetworkType string
	Address     string
}

type Listener struct {
	Shutdown chan bool
	l        net.Listener
}

func New(logWriter io.Writer, dbPath string) *Server {
	return &Server{
		Shutdown:      make(chan bool),
		SignalHandler: make(chan os.Signal),
		Logger:        log.New(logWriter, "[NETWORK] ", log.LstdFlags),
		Storage:       storage.New(logWriter, dbPath),
		listeners:     make([]*Listener, 0),
		logWriter:     logWriter,
	}
}

func NewListener(l net.Listener) *Listener {
	return &Listener{
		Shutdown: make(chan bool),
		l:        l,
	}
}

func (s *Server) Start(addresses []ListenAddress) {
	done := false
	s.Logger.Printf("Starting...\n")
	// Open network listeners
	for _, address := range addresses {
		l, err := s.Listen(address)
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

func (s *Server) Listen(address ListenAddress) (*Listener, error) {
	l, err := net.Listen(address.NetworkType, address.Address)
	if err != nil {
		return nil, err
	}
	s.Logger.Printf("Listening on %s %s.\n", address.NetworkType, address.Address)
	listener := NewListener(l)
	go s.acceptConnections(listener)
	return listener, nil
}

func (s *Server) acceptConnections(l *Listener) {
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

	enc := encoder.NewEncoder(c)
	dec := encoder.NewDecoder(c)

	for {
		command := Command{}
		response := Response{}
		err := dec.Decode(&command)
		if err == io.EOF {
			// Done
			break
		} else if err != nil {
			response.Error = err.Error()
		} else {
			id := storage.IDType(command.ID)
			value := storage.StorageType(command.Value)
			switch command.Type {
			case TimeCommand:
				response.Value = fmt.Sprintf("Hello! The time is currently %s.\n",
					t.Format(time.RFC3339))
			case GetCommand:
				v, err := s.get(id)
				response.ID = id.String()
				response.Value = v.String()
				if err != nil {
					response.Error = err.Error()
				}
			case UpdateCommand:
				v, err := s.update(id, value)
				response.ID = id.String()
				response.Value = v.String()
				if err != nil {
					response.Error = err.Error()
				}
			}
		}
		enc.Encode(&response)
	}

	c.Close()
}

func (s *Server) get(id storage.IDType) (storage.StorageType, error) {
	request := storage.GetRequest{
		ID:     id,
		Result: make(chan storage.Result),
	}
	s.Storage.Get <- request
	result := <-request.Result
	return result.Value, result.Error
}

func (s *Server) update(id storage.IDType, value storage.StorageType) (storage.StorageType, error) {
	request := storage.UpdateRequest{
		ID:     id,
		Value:  value,
		Result: make(chan storage.Result),
	}
	s.Storage.Update <- request
	result := <-request.Result
	return result.Value, result.Error
}
