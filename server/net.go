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

// CommandType represents the type of a server command
type CommandType uint8

const (
	// TimeCommand returns the server's current timestamp.
	TimeCommand CommandType = 1
	// GetCommand returns a value
	GetCommand CommandType = 2
	// UpdateCommand updates a value
	UpdateCommand CommandType = 3
	// RemoveCommand removes a value
	RemoveCommand CommandType = 4
)

// Command is a server command
type Command struct {
	Type  CommandType
	ID    string
	Value string
}

// Response is returned from a server command
type Response struct {
	ID    string
	Value string
	Error string
}

func (r *Response) String() string {
	return fmt.Sprintf("Response [ID: %X, Value: %s, Error: %s]", r.ID, r.Value, r.Error)
}

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
			id := command.ID
			value := command.Value
			switch command.Type {
			case TimeCommand:
				response.Value = fmt.Sprintf("Hello! The time is currently %s.\n",
					t.Format(time.RFC3339))
			case GetCommand:
				v := s.get(id)
				response.ID = id
				response.Value = v
			case UpdateCommand:
				v := s.update(id, value)
				response.ID = id
				response.Value = v
			}
		}
		enc.Encode(&response)
	}

	c.Close()
}

func (s *Server) get(id string) string {
	request := storage.GetRequest{
		ID:     id,
		Remove: false,
		Result: make(chan storage.Result),
	}
	s.Storage.Get <- request
	result := <-request.Result
	return result.Value
}

func (s *Server) update(id string, value string) string {
	request := storage.UpdateRequest{
		ID:     id,
		Value:  value,
		Result: make(chan storage.Result),
	}
	s.Storage.Update <- request
	result := <-request.Result
	return result.Value
}

func (s *Server) remove(id string) string {
	request := storage.GetRequest{
		ID:     id,
		Remove: true,
		Result: make(chan storage.Result),
	}
	s.Storage.Get <- request
	result := <-request.Result
	return result.Value
}
