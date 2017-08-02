package storage

import (
	"io"
	"log"
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
	Id IdType
	Result chan Result
}

type UpdateRequest struct {
	Id IdType
	Value StorageType
	Result chan Result
}

type Result struct {
	Error error
	Id IdType
	Value StorageType
}

type Storage struct {
	Get chan GetRequest
	Update chan UpdateRequest
	Shutdown chan bool
	Logger *log.Logger
	storage map[IdType]StorageType
}

func New(logWriter io.Writer) *Storage {
	db := &Storage {
		Get: make(chan GetRequest),
		Update: make(chan UpdateRequest),
		Shutdown: make(chan bool),
		Logger: log.New(logWriter, "[STORAGE] ", log.LstdFlags),
		storage: make(map[IdType]StorageType),
	}
	go db.Start()
	return db
}

func (db *Storage) Start() {
	db.Logger.Printf("Started\n")
	done := false
	for {
		select {
		case get := <-db.Get:
			value, error := db.getValue(get.Id)
			result := Result {
				Error: error,
				Id: get.Id,
				Value: value,
			}
			get.Result <- result
		case update := <-db.Update:
			value, error := db.setValue(update.Id, update.Value)
			result := Result {
				Error: error,
				Id: update.Id,
				Value: value,
			}
			update.Result <- result
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
