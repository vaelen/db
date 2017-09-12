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

package storage

import (
	"io"
	"log"
	"os"

	"encoding/gob"
	"path/filepath"
)

type IdType string
type StorageType string

func (id IdType) String() string {
	return string(id)
}

func (value StorageType) String() string {
	return string(value)
}

type GetRequest struct {
	Id     IdType
	Result chan Result
}

type UpdateRequest struct {
	Id     IdType
	Value  StorageType
	Result chan Result
}

type Result struct {
	Error error
	Id    IdType
	Value StorageType
}

type Storage struct {
	Get      chan GetRequest
	Update   chan UpdateRequest
	Shutdown chan bool
	Logger   *log.Logger
	Path     string
	storage  map[IdType]StorageType
	dirty    bool
}

func New(logWriter io.Writer, dbPath string) *Storage {
	db := &Storage{
		Get:      make(chan GetRequest),
		Update:   make(chan UpdateRequest),
		Shutdown: make(chan bool),
		Logger:   log.New(logWriter, "[STORAGE] ", log.LstdFlags),
		Path:     dbPath,
		storage:  make(map[IdType]StorageType),
	}
	db.load()
	go db.Start()
	return db
}

func (db *Storage) Start() {
	db.Logger.Printf("Started: %s\n", db.Path)
	done := false
	for {
		select {
		case get := <-db.Get:
			value, error := db.getValue(get.Id)
			result := Result{
				Error: error,
				Id:    get.Id,
				Value: value,
			}
			get.Result <- result
		case update := <-db.Update:
			value, error := db.setValue(update.Id, update.Value)
			result := Result{
				Error: error,
				Id:    update.Id,
				Value: value,
			}
			update.Result <- result
			db.dirty = true
		case <-db.Shutdown:
			done = true
			db.Logger.Printf("Stopping...\n")
		default:
			if db.dirty {
				db.save()
				db.dirty = false
			}
		}
		if done {
			break
		}
	}
	db.Logger.Printf("Stopped\n")
}

func (db *Storage) save() {
	filename := filepath.Join(db.Path, "storage.gob")
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		db.Logger.Fatalf("Could not open file for saving. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(&db.storage)
	if err != nil {
		db.Logger.Fatalf("Could not save storage. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	db.Logger.Printf("Storage saved: %s\n", filename)
}

func (db *Storage) load() {
	filename := filepath.Join(db.Path, "storage.gob")
	f, err := os.OpenFile(filename, os.O_RDONLY, 0640)
	if err != nil {
		db.Logger.Printf("Could not open file for loading. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	err = dec.Decode(&db.storage)
	if err != nil {
		db.Logger.Fatalf("Could not load storage. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	db.Logger.Printf("Storage loaded: %s\n", filename)
}

func (db *Storage) getValue(id IdType) (StorageType, error) {
	value, ok := db.storage[id]
	if !ok {
		value = ""
	}
	db.Logger.Printf("Get - Key: %s, Value: %s\n", id, value)
	return value, nil
}

func (db *Storage) setValue(id IdType, value StorageType) (StorageType, error) {
	db.storage[id] = value
	db.Logger.Printf("Update - Key: %s, Value: %s\n", id, value)
	return value, nil
}
