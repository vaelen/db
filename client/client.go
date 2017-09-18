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
	encoder "encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/vaelen/db/server"
)

// Client is an instance of the database client
type Client struct {
	Logger *log.Logger
	conn   net.Conn
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
	return nil
}

// Command excutes a database command on the server
func (c *Client) Command(command server.Command) server.Response {
	response := server.Response{}

	if c.conn == nil {
		c.Logger.Printf("Error - Not Connected\n")
		response.Error = "Not Connected"
		return response
	}

	enc := encoder.NewEncoder(c.conn)
	dec := encoder.NewDecoder(c.conn)

	err := enc.Encode(&command)
	if err != nil {
		c.Logger.Printf("Encoding error: %s\n", err.Error())
		response.Error = err.Error()
		return response
	}

	err = dec.Decode(&response)
	if err != nil {
		c.Logger.Printf("Decoding error: %s\n", err.Error())
		response.Error = err.Error()
		return response
	}

	return response
}

// Time returns the server's current timestamp
func (c *Client) Time() (string, error) {
	command := server.Command{
		Type: server.TimeCommand,
	}
	response := c.Command(command)
	if response.Error != "" {
		return "", fmt.Errorf(response.Error)
	}
	return response.Value, nil
}

// Get returns a value from the server
func (c *Client) Get(id string) (string, error) {
	command := server.Command{
		Type: server.GetCommand,
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
	command := server.Command{
		Type:  server.UpdateCommand,
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
	command := server.Command{
		Type: server.RemoveCommand,
		ID:   id,
	}
	response := c.Command(command)
	if response.Error != "" {
		return "", fmt.Errorf(response.Error)
	}
	c.Logger.Printf("Remove - %s\n", response.String())
	return response.Value, nil
}
