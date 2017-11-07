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
	"os"
	"github.com/vaelen/db/server"
	"testing"
	"fmt"
	"net"
	"log"
	"google.golang.org/grpc"
	"github.com/vaelen/db/api"
)

func TestDBClient(t *testing.T) {
	testPort := 30000
	address := fmt.Sprintf("localhost:%d", testPort)

	s := server.New(os.Stdout, "")
	defer func() { s.Stop() }()

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterDatabaseServer(grpcServer, s)
	go grpcServer.Serve(lis)
	defer func() { grpcServer.Stop() }()

	c := New(os.Stderr)
	err = c.Connect(address)
	if err != nil {
		t.Fatalf("Connect Error: %s\n", err.Error())
	}

	id := "foo"
	value := "bar"

	ct, err := c.Time()
	if err != nil {
		t.Fatalf("Time Error: %s\n", err.Error())
	}
	t.Logf("Time - %s\n", ct)

	err = c.Set(id, value)
	if err != nil {
		t.Fatalf("Set Error: %s\n", err.Error())
	}
	t.Logf("Set - Key: %s, Value: %s\n", id, value)

	newValue, err := c.Get(id)
	if err != nil {
		t.Fatalf("Get Error: %s\n", err.Error())
	}
	t.Logf("Get - Key: %s, Value: %s\n", id, newValue)

	oldValue, err := c.Remove(id)
	if err != nil {
		t.Fatalf("Remove Error: %s\n", err.Error())
	}
	t.Logf("Remove - Key: %s, Value: %s\n", id, oldValue)
}