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

// Package storage provides a thread safe hash table implementation.
// The hash table is implemented internally as a tree structure and is self-pruning.
// The hash table uses a 32-bit FNV hash and each node of the tree represents 1 byte in the hash.
// The intention is to allow movement of entire nodes of the tree between storage instances.
package storage

import (
	"io"
	"log"
	"os"

	"encoding/gob"
	"path/filepath"
)

// GetRequest is used to retrieve a value and optionally remove it from the storage tree.
type GetRequest struct {
	ID     string
	Remove bool
	Result chan Result
}

// SetRequest is used to set a value in the storage tree.
type SetRequest struct {
	ID     string
	Value  string
	Result chan Result
}

// Result is returned from Get and Set
type Result struct {
	ID    string
	Value string
}

// GetNodeRequest is used to request an entire node of the storage tree.
type GetNodeRequest struct {
	ID     NodeLocator
	Remove bool
	Result chan NodeResult
}

// SetNodeRequest is used to set an entire node of the storage tree.
type SetNodeRequest struct {
	ID     NodeLocator
	Value  *Node
	Result chan NodeResult
}

// NodeResult is returned from GetNode and SetNode
type NodeResult struct {
	ID    NodeLocator
	Value *Node
}

// Instance represents a storage tree instance
type Instance struct {
	// Get retrieves a value from storage, optionally removing it
	getChannel chan GetRequest
	// Set sets a value in storage
	setChannel chan SetRequest
	// GetNode retrieves a node from the storage tree, optionally removing it
	GetNode chan GetNodeRequest
	// SetNode sets a node in the storage tree
	SetNode chan SetNodeRequest
	// Shutdown stops the storage worker thread
	Shutdown chan bool
	// Logger is the logger instance used by the storage instance
	Logger *log.Logger
	// Path is the path to the data file maintained by this storage instance
	Path    string
	storage *Hashtable
}

// New creates a new Storage instance. It also loads the data file if it exists and starts the storage thread.
func New(logWriter io.Writer, dbPath string) *Instance {
	db := &Instance{
		getChannel: make(chan GetRequest),
		setChannel: make(chan SetRequest),
		GetNode:    make(chan GetNodeRequest),
		SetNode:    make(chan SetNodeRequest),
		Shutdown:   make(chan bool),
		Logger:     log.New(logWriter, "[STORAGE] ", log.LstdFlags),
		Path:       dbPath,
		storage:    NewHashtable(),
	}
	db.load()
	go db.start()
	return db
}

func (db *Instance) start() {
	db.Logger.Printf("Started: %s\n", db.Path)
	done := false
	for {
		select {
		case get := <-db.getChannel:
			value := db.storage.Get(get.ID)
			if get.Remove {
				db.storage.Remove(get.ID)
			}
			result := Result{
				ID:    get.ID,
				Value: value,
			}
			//db.Logger.Printf("Get - Key: %s, Value: %s\n", get.ID, value)
			get.Result <- result
		case set := <-db.setChannel:
			db.storage.Set(set.ID, set.Value)
			result := Result{
				ID:    set.ID,
				Value: set.Value,
			}
			//db.Logger.Printf("Set - Key: %s, Value: %s\n", set.ID, set.Value)
			set.Result <- result
			db.save()
		case getNode := <-db.GetNode:
			node, _ := db.storage.FindNode(getNode.ID)
			if node != nil && getNode.Remove {
				db.storage.RemoveNode(getNode.ID)
			}
			result := NodeResult{
				ID:    getNode.ID,
				Value: node,
			}
			getNode.Result <- result
		case setNode := <-db.SetNode:
			node := db.storage.SetNode(setNode.ID, setNode.Value)
			result := NodeResult{
				ID:    setNode.ID,
				Value: node,
			}
			setNode.Result <- result
			db.save()
		case <-db.Shutdown:
			done = true
			db.Logger.Printf("Stopping...\n")
		}
		if done {
			break
		}
	}
	db.Logger.Printf("Stopped\n")
}

func (db *Instance) save() {
	if true {
		// TODO: Implement custom save/load routine
		return
	}
	if db.Path == "" {
		return
	}
	filename := filepath.Join(db.Path, "storage.gob")
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		db.Logger.Fatalf("Could not open file for saving. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(db.storage)
	if err != nil {
		db.Logger.Fatalf("Could not save storage. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	db.Logger.Printf("Storage saved: %s\n", filename)
}

func (db *Instance) load() {
	if db.Path == "" {
		return
	}

	filename := filepath.Join(db.Path, "storage.gob")
	f, err := os.OpenFile(filename, os.O_RDONLY, 0640)
	if err != nil {
		db.Logger.Printf("Could not open file for loading. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	err = dec.Decode(db.storage)
	if err != nil {
		db.Logger.Fatalf("Could not load storage. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	db.Logger.Printf("Storage loaded: %s\n", filename)
}

// Get returns the value of the given key
func (db *Instance) Get(id string) string {
	request := GetRequest{
		ID:     id,
		Remove: false,
		Result: make(chan Result),
	}
	db.getChannel <- request
	result := <-request.Result
	return result.Value
}

// Set sets the value of the given key
func (db *Instance) Set(id string, value string) string {
	request := SetRequest{
		ID:     id,
		Value:  value,
		Result: make(chan Result),
	}
	db.setChannel <- request
	result := <-request.Result
	return result.Value
}

// Remove removes the given key
func (db *Instance) Remove(id string) string {
	request := GetRequest{
		ID:     id,
		Remove: true,
		Result: make(chan Result),
	}
	db.getChannel <- request
	result := <-request.Result
	return result.Value
}

// Close shuts down the storage instance
func (db *Instance) Close() {
	db.Shutdown <- true
}
