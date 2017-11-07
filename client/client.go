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

package client

import (
	"io"
	"log"

	"github.com/vaelen/db/api"
	"google.golang.org/grpc"
	"context"
)

// DBClient is an instance of the database client
type DBClient struct {
	Logger *log.Logger
	conn   *grpc.ClientConn
	client api.DatabaseClient
}

// New creates a new DBClient instance
func New(logWriter io.Writer) *DBClient {
	return &DBClient{
		Logger: log.New(logWriter, "[CLIENT] ", log.LstdFlags),
	}
}

// Connect to a database server
func (c *DBClient) Connect(address string) error {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.conn = conn
	c.client = api.NewDatabaseClient(c.conn)
	return nil
}

// Close disconnects from database server
func (c *DBClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// Time returns the server's current timestamp
func (c *DBClient) Time() (string, error) {
	response, err := c.client.Time(context.Background(), &api.EmptyRequest{})
	return response.Value, err
}

// Get returns a value from the server
func (c *DBClient) Get(id string) (string, error) {
	response, err := c.client.Get(context.Background(), &api.IDRequest{ ID: id })
	return response.Value, err
}

// Set sets a value on the server
func (c *DBClient) Set(id string, value string) error {
	_, err := c.client.Set(context.Background(), &api.IDValueRequest{ ID: id, Value: value })
	return err
}

// Remove removes a value from the database server
func (c *DBClient) Remove(id string) (string, error) {
	response, err := c.client.Remove(context.Background(), &api.IDRequest{ ID: id })
	return response.Value, err
}


