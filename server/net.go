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
	Type CommandType
	Id string
	Value string
}

type Response struct {
	Id string
	Value string
	Error string
}

type Server struct {
	Shutdown chan bool
	SignalHandler chan os.Signal
	Logger *log.Logger
	Storage *storage.Storage
	listeners []*Listener
	logWriter io.Writer
}

type ListenAddress struct {
	NetworkType string
	Address string
}

type Listener struct {
	Shutdown chan bool
	l net.Listener
}

func New(logWriter io.Writer, dbPath string) *Server {
	return &Server{
		Shutdown: make(chan bool),
		SignalHandler: make(chan os.Signal),
		Logger: log.New(logWriter, "[NETWORK] ", log.LstdFlags),
		Storage: storage.New(logWriter, dbPath),
		listeners: make([]*Listener, 0),
		logWriter: logWriter,
	}
}

func NewListener(l net.Listener) *Listener {
	return &Listener {
		Shutdown: make(chan bool),
		l: l,
	}
}

func (s *Server) Start(addresses []ListenAddress) {
	done := false
	s.Logger.Printf("Starting...\n")
	// Open network listeners
	for _, address := range addresses {
		l, err := s.Listen(address)
		if (err != nil) {
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
		case <- l.Shutdown:
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
			id := storage.IdType(command.Id)
			value := storage.StorageType(command.Value)
			switch command.Type {
			case TimeCommand:
				response.Value = fmt.Sprintf("Hello! The time is currently %s.\n",
					t.Format(time.RFC3339))
			case GetCommand:
				v, err := s.get(id)
				response.Id = id.String()
				response.Value = v.String()
				if err != nil {
					response.Error = err.Error()
				}
			case UpdateCommand:
				v, err := s.update(id, value)
				response.Id = id.String()
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

func (s *Server) get(id storage.IdType) (storage.StorageType, error) {
	request := storage.GetRequest {
		Id: id,
		Result: make(chan storage.Result),
	}
	s.Storage.Get <- request
	result := <-request.Result
	return result.Value, result.Error
}

func (s *Server) update(id storage.IdType, value storage.StorageType) (storage.StorageType, error) {
	request := storage.UpdateRequest {
		Id: id,
		Value: value,
		Result: make(chan storage.Result),
	}
	s.Storage.Update <- request
	result := <-request.Result
	return result.Value, result.Error
}
