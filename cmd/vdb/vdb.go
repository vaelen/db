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
	"log"
	"os"
	"math/rand"
	"time"

	"github.com/vaelen/db/client"
	"github.com/vaelen/db/server"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

func main() {
	c := client.New(os.Stderr)
	err := c.Connect(server.ListenAddress{NetworkType: "tcp", Address: "localhost:5555"})
	if err != nil {
		log.Fatalf("Connect Error: %s\n", err.Error())
	}

	id := "foo"
	value := "bar"

	ct, err := c.Time()
	if err != nil {
		log.Fatalf("Time Error: %s\n", err.Error())
	}
	log.Printf("Time - %s\n", ct)

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

	oldValue, err := c.Remove(id)
	if err != nil {
		log.Fatalf("Remove Error: %s\n", err.Error())
	}
	log.Printf("Remove - Key: %s, Value: %s\n", id, oldValue)

	// Generate some random data
	rand.Seed(time.Now().UnixNano())

	log.Printf("Generating random strings\n")
	
	m := make(map[string]string)
	for i := 0; i < 100000; i++ {
		m[randomString(1000)] = randomString(1000)
	}

	log.Printf("Sending random strings\n")
	for k, v := range m {
		c.Update(k, v)
	}

	log.Printf("Getting random strings\n")
	failures := 0
	for k, v := range m {
		x, _ := c.Get(k)
		if x != v {
			failures++
			log.Printf("Error Getting Value: Key: %s\n", k)
		}
	}

	log.Printf("Removing random strings\n")
	for k := range m {
		_, _ = c.Remove(k)
	}

	log.Printf("%d failures\n", failures)


}
