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
