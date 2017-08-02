package main

import (
	"os"
	"log"
	"github.com/vaelen/db/server"
	"github.com/vaelen/db/client"
)

func main() {
	c := client.New(os.Stderr)
	err := c.Connect(server.ListenAddress{ NetworkType: "tcp", Address: "localhost:5555" })
	if err != nil {
		log.Fatalf("Connect Error: %s\n", err.Error())
	}

	id := "foo"
	value := "bar"
	
	time, err := c.Time()
	if err != nil {
		log.Fatalf("Time Error: %s\n", err.Error())
	}
	log.Printf("Time - %s\n", time)

	err = c.Update(id, value)
	if err != nil {
		log.Fatalf("Update Error: %s\n", err.Error())
	}
	log.Printf("Set - Key: %s, Value: %s\n", id, value)

	newValue, err := c.Get(id)
	if err != nil {
		log.Fatalf("Get Error: %s\n", err.Error())
	}
	log.Printf("Get - Key: %s, Value: %s\n", id, newValue)

}
