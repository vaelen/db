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
	"io"
	"log"
	"time"

	"github.com/vaelen/db/storage"
	"github.com/vaelen/db/api"

	"golang.org/x/net/context"
)

// DBServer is an instance of the database server
type DBServer struct {
	Logger        *log.Logger
	Storage       *storage.Instance
	logWriter     io.Writer
}

// New creates a new instance of the database server
func New(logWriter io.Writer, dbPath string) *DBServer {
	return &DBServer{
		Logger:        log.New(logWriter, "[NETWORK] ", log.LstdFlags),
		Storage:       storage.New(logWriter, dbPath),
		logWriter:     logWriter,
	}
}

// Stop shuts down the database server
func (s *DBServer) Stop() {
	if s.Storage != nil {
		s.Storage.Shutdown <- true
	}
}

// Time returns the current time
func (s *DBServer) Time(ctx context.Context, request *api.EmptyRequest) (*api.Response, error) {
	return &api.Response{
		Value: time.Now().Format(time.RFC3339),
	}, nil
}

// Get returns a value for a given key
func (s *DBServer) Get(ctx context.Context, request *api.IDRequest) (*api.Response, error) {
	return &api.Response{
		Value: s.Storage.Get(request.ID),
	}, nil
}

// Set sets a value for a given key
func (s *DBServer) Set(ctx context.Context, request *api.IDValueRequest) (*api.Response, error) {
	return &api.Response{
		Value: s.Storage.Set(request.ID, request.Value),
	}, nil
}

// Remove removes a given key
func (s *DBServer) Remove(ctx context.Context, request *api.IDRequest) (*api.Response, error) {
	return &api.Response{
		Value: s.Storage.Remove(request.ID),
	}, nil
}


