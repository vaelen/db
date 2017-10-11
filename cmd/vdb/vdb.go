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

package main

import (
	"os"

	"github.com/abiosoft/ishell"

	"github.com/vaelen/db/client"
	"github.com/vaelen/db/server"
)

func Start() {
	db := client.New(os.Stderr)
	defer db.Close()

	shell := ishell.New()

	// display welcome info.
	shell.Println("Vaelen/DB Client v0.1")

	// register a function for "greet" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "connect",
		Help: "connects to a Vaelen/DB server. usage: connect <address>",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("Usage: connect <address>")
				return
			}
			shell.Printf("Connecting to %s...\n", c.Args[0])
			err := db.Connect(server.ListenAddress{NetworkType: "tcp", Address: c.Args[0]})
			if err != nil {
				c.Printf("Error: %s\n", err)
				return
			}
			c.Println("Connected")
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "time",
		Help: "returns the current server time. usage: time",
		Func: func(c *ishell.Context) {
			ct, err := db.Time()
			if err != nil {
				c.Printf("Error: %s\n", err)
				return
			}
			c.Printf("Server Time: %s\n", ct)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "returns the value for a given key. usage: get <key>",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("Usage: get <key>")
				return
			}
			v, err := db.Get(c.Args[0])
			if err != nil {
				c.Printf("Error: %s\n", err)
				return
			}
			c.Printf("Value: %s\n", v)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "sets the value for a given key. usage: get <key> <value>",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Println("Usage: set <key> <value>")
				return
			}
			err := db.Set(c.Args[0], c.Args[1])
			if err != nil {
				c.Printf("Error: %s\n", err)
				return
			}
			c.Println("Value Set")
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "remove",
		Help: "removes the value for a given key. usage: remove <key>",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("Usage: remove <key>")
				return
			}
			_, err := db.Remove(c.Args[0])
			if err != nil {
				c.Printf("Error: %s\n", err)
				return
			}
			c.Println("Value Removed")
		},
	})

	// initial connection
	address := "localhost:5555"
	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	shell.Printf("Connecting to %s...\n", address)
	err := db.Connect(server.ListenAddress{NetworkType: "tcp", Address: address})
	if err != nil {
		shell.Printf("Error: %s\n", err.Error())
	}
	shell.Println("Connected")

	// run shell
	shell.Run()
}

func main() {
	Start()
}
