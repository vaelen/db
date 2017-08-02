package client

import (
	encoder "encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/vaelen/db/server"
)

type Client struct {
	Logger *log.Logger
	conn net.Conn
}

func New(logWriter io.Writer) *Client {
	return &Client{
		Logger: log.New(logWriter, "[CLIENT] ", log.LstdFlags),
	}
}

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
	if (err != nil) {
		c.Logger.Printf("Encoding error: %s\n", err.Error())
		response.Error = err.Error()
		return response
	}

	err = dec.Decode(&response)
	if (err != nil) {
		c.Logger.Printf("Decoding error: %s\n", err.Error())
		response.Error = err.Error()
		return response
	}
	
	return response
}

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

func (c *Client) Get(id string) (string, error) {
	command := server.Command{
		Type: server.GetCommand,
		Id: id,
	}
	response := c.Command(command)
	if response.Error != "" {
		return "", fmt.Errorf(response.Error)
	}
	c.Logger.Printf("Get - %s\n", response)
	return response.Value, nil
}

func (c *Client) Update(id string, value string) error {
	command := server.Command{
		Type: server.UpdateCommand,
		Id: id,
		Value: value,
	}
	response := c.Command(command)
	if response.Error != "" {
		return fmt.Errorf(response.Error)
	}
	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
