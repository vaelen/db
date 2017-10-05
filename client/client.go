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
	"fmt"
	"io"
	"log"
	"net"

	"github.com/vaelen/db/server"
	"github.com/vaelen/db/api"
	"github.com/golang/protobuf/proto"
)

// Client is an instance of the database client
type Client struct {
	Logger *log.Logger
	conn   net.Conn
	m api.Multiplexer
}

// New creates a new Client instance
func New(logWriter io.Writer) *Client {
	return &Client{
		Logger: log.New(logWriter, "[CLIENT] ", log.LstdFlags),
	}
}

// Connect to a database server
func (c *Client) Connect(address server.ListenAddress) error {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	conn, err := net.Dial(address.NetworkType, address.Address)
	if err != nil {
		return nil
	}
	c.conn = conn
	c.m = api.Multiplexer{ Stream: conn }
	return nil
}

// Command executes a database command on the server
func (c *Client) Command(command *api.Command) *api.Response {
	response := &api.Response{}

	if c.conn == nil {
		c.Logger.Printf("Error - Not Connected\n")
		response.Error = "Not Connected"
		return response
	}

	buf, err := proto.Marshal(command)
	if err != nil {
		c.Logger.Printf("Encoding error: %s\n", err.Error())
		response.Error = err.Error()
		return response
	}
	err = c.m.Send(buf)
	if err != nil {
		c.Logger.Printf("Sending error: %s\n", err.Error())
		response.Error = err.Error()
		return response
	}

	buf, err = c.m.Receive()
	if err != nil {
		c.Logger.Printf("Receiving error: %s\n", err.Error())
		response.Error = err.Error()
		return response
	}

	err = proto.Unmarshal(buf, response)
	if err != nil {
		c.Logger.Printf("Decoding error: %s\n", err.Error())
		response.Error = err.Error()
		return response
	}

	return response
}

// Time returns the server's current timestamp
func (c *Client) Time() (string, error) {
	command := &api.Command{
		Type: api.Command_TIME,
	}
	response := c.Command(command)
	if response.Error != "" {
		return "", fmt.Errorf(response.Error)
	}
	return response.Value, nil
}

// Get returns a value from the server
func (c *Client) Get(id string) (string, error) {
	command := &api.Command{
		Type: api.Command_GET,
		ID:   id,
	}
	response := c.Command(command)
	if response.Error != "" {
		return "", fmt.Errorf(response.Error)
	}
	c.Logger.Printf("Get - %s\n", response.String())
	return response.Value, nil
}

// Update sets a value on the server
func (c *Client) Update(id string, value string) error {
	command := &api.Command{
		Type:  api.Command_SET,
		ID:    id,
		Value: value,
	}
	response := c.Command(command)
	if response.Error != "" {
		return fmt.Errorf(response.Error)
	}
	return nil
}

// Close disconnects from database server
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// Remove removes a value from the database server
func (c *Client) Remove(id string) (string, error) {
	command := &api.Command{
		Type: api.Command_REMOVE,
		ID:   id,
	}
	response := c.Command(command)
	if response.Error != "" {
		return "", fmt.Errorf(response.Error)
	}
	c.Logger.Printf("Remove - %s\n", response.String())
	return response.Value, nil
}
