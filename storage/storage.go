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
	"bytes"
	"io"
	"log"
	"os"

	"encoding/binary"
	"encoding/gob"
	"hash/fnv"
	"path/filepath"
)

// GetRequest is used to retrieve a value and optionally remove it from the storage tree.
type GetRequest struct {
	ID     string
	Remove bool
	Result chan Result
}

// SetRequest is used to update a value in the storage tree.
type SetRequest struct {
	ID     string
	Value  string
	Result chan Result
}

// Result is returned from Get and Update
type Result struct {
	ID    string
	Value string
}

// NodeLocator provides an address to a specific node in the storage tree.
type NodeLocator struct {
	ID    uint32
	Bytes byte
}

// GetNodeLocator returns a NodeLocator for this ID
func GetNodeLocator(id string) NodeLocator {
	return NodeLocator{
		ID:    Hash(id),
		Bytes: 4,
	}
}

// GetBytes returns a byte slice representing this node locator
func (id NodeLocator) GetBytes() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, id)
	if err != nil {
		log.Fatalf("Couldn't convert number to byte array: %s\n", err.Error())
	}
	b := buf.Bytes()
	return b[:id.Bytes]
}

// GetNodeRequest is used to request an entire node of the storage tree.
type GetNodeRequest struct {
	ID     NodeLocator
	Remove bool
	Result chan NodeResult
}

// UpdateNodeRequest is used to update an entire node of the storage tree.
type UpdateNodeRequest struct {
	ID     NodeLocator
	Value  *Node
	Result chan NodeResult
}

// NodeResult is returned from GetNode and UpdateNode
type NodeResult struct {
	ID    NodeLocator
	Value *Node
}

// Node represents a single node in the storage tree
type Node struct {
	Children [256]*Node
	values   []NodeKeyValuePair
}

// NewNode returns a new Node instance
func NewNode() *Node {
	return &Node{
		values: make([]NodeKeyValuePair, 0),
	}
}

// NodeKeyValuePair holds a key/value pair
type NodeKeyValuePair struct {
	Key   string
	Value string
}

// IsLeaf returns true if this node is a leaf node
func (n *Node) IsLeaf() bool {
	return len(n.values) > 0
}

// GetValue returns the given value from the node
func (n *Node) GetValue(key string) string {
	for _, v := range n.values {
		if v.Key == key {
			return v.Value
		}
	}
	var value string
	return value
}

// SetValue sets the given value on the node
func (n *Node) SetValue(key string, value string) {
	n.RemoveValue(key)
	n.values = append(n.values, NodeKeyValuePair{Key: key, Value: value})
}

// RemoveValue removes the given value from the node
func (n *Node) RemoveValue(key string) {
	for i, v := range n.values {
		if v.Key == key {
			n.values = append(n.values[:i], n.values[i+1:]...)
		}
	}
}

// IsEmpty returns true if this node is empty.  Leaf nodes are empty if they have no value.  Non-leaf nodes are empty if they have no children.
func (n *Node) IsEmpty() bool {
	found := false
	for _, x := range n.Children {
		if x != nil {
			found = true
			break
		}
	}
	return !n.IsLeaf() && !found
}

// Storage represents a storage tree instance
type Storage struct {
	// Get retrieves a value from storage, optionally removing it
	getChannel chan GetRequest
	// Update updates a value in storage
	setChannel chan SetRequest
	// GetNode retrieves a node from the storage tree, optionally removing it
	GetNode chan GetNodeRequest
	// UpdateNode updates a node in the storage tree
	UpdateNode chan UpdateNodeRequest
	// Shutdown stops the storage worker thread
	Shutdown chan bool
	// Logger is the logger instance used by the storage instance
	Logger *log.Logger
	// Path is the path to the data file maintained by this storage instance
	Path    string
	storage *Node
}

// New creates a new Storage instance. It also loads the data file if it exists and starts the storage thread.
func New(logWriter io.Writer, dbPath string) *Storage {
	db := &Storage{
		getChannel: make(chan GetRequest),
		setChannel: make(chan SetRequest),
		GetNode:    make(chan GetNodeRequest),
		UpdateNode: make(chan UpdateNodeRequest),
		Shutdown:   make(chan bool),
		Logger:     log.New(logWriter, "[STORAGE] ", log.LstdFlags),
		Path:       dbPath,
		storage:    NewNode(),
	}
	db.load()
	go db.start()
	return db
}

func (db *Storage) start() {
	db.Logger.Printf("Started: %s\n", db.Path)
	done := false
	for {
		select {
		case get := <-db.getChannel:
			id := GetNodeLocator(get.ID)
			node, path := db.findNode(id)
			var value string
			if node != nil {
				value = node.GetValue(get.ID)
				if get.Remove {
					node.RemoveValue(get.ID)
					db.prune(id.GetBytes(), path)
				}
			}
			result := Result{
				ID:    get.ID,
				Value: value,
			}
			//db.Logger.Printf("Get - Key: %s, Value: %s\n", get.ID, value)
			get.Result <- result
		case set := <-db.setChannel:
			id := GetNodeLocator(set.ID)
			node := db.setNodeValue(id, set.ID, set.Value)
			if node != nil {
				node.SetValue(set.ID, set.Value)
			}
			result := Result{
				ID:    set.ID,
				Value: set.Value,
			}
			//db.Logger.Printf("Set - Key: %s, Value: %s\n", set.ID, set.Value)
			set.Result <- result
			db.save()
		case getNode := <-db.GetNode:
			node, _ := db.findNode(getNode.ID)
			if node != nil && getNode.Remove {
				db.removeNode(getNode.ID)
			}
			result := NodeResult{
				ID:    getNode.ID,
				Value: node,
			}
			getNode.Result <- result
		case updateNode := <-db.UpdateNode:
			node := db.setNode(updateNode.ID, updateNode.Value)
			result := NodeResult{
				ID:    updateNode.ID,
				Value: node,
			}
			updateNode.Result <- result
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

func (db *Storage) save() {
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
	err = enc.Encode(&db.storage)
	if err != nil {
		db.Logger.Fatalf("Could not save storage. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	db.Logger.Printf("Storage saved: %s\n", filename)
}

func (db *Storage) load() {
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
	err = dec.Decode(&db.storage)
	if err != nil {
		db.Logger.Fatalf("Could not load storage. File: %s, Error: %s\n", filename, err.Error())
		return
	}
	db.Logger.Printf("Storage loaded: %s\n", filename)
}

// findNode returns a node and the path to its parent node
func (db *Storage) findNode(id NodeLocator) (*Node, []*Node) {
	path := make([]*Node, 0)
	return db.findNodeRecurse(id.GetBytes(), db.storage, path)
}

// findNodeRecurse returns a node and the path to its parent node
func (db *Storage) findNodeRecurse(id []byte, parent *Node, path []*Node) (*Node, []*Node) {
	if parent == nil {
		// We've hit a deadend
		return nil, path
	}

	if len(id) == 0 {
		// We found it
		return parent, path
	}

	// Not yet found, recurse
	return db.findNodeRecurse(id[1:], parent.Children[id[0]], append(path, parent))
}

func (db *Storage) setNodeValue(id NodeLocator, key string, value string) *Node {
	node, _ := db.findNode(id)
	if node != nil {
		node.SetValue(key, value)
	} else {
		node = NewNode()
		node.SetValue(key, value)
		db.setNode(id, node)
	}
	return node
}

func (db *Storage) setNode(id NodeLocator, value *Node) *Node {
	return db.setNodeRecurse(id.GetBytes(), db.storage, value)
}

func (db *Storage) setNodeRecurse(id []byte, parent *Node, child *Node) *Node {
	if parent == nil || len(id) == 0 {
		// This should never happen, but just in case...
		db.Logger.Printf("WARNING: setNodeRecurse() called with either no parent or no id.\n")
		return nil
	}

	if len(id) == 1 {
		// We need to set the child on this node
		parent.Children[id[0]] = child
		return parent.Children[id[0]]
	}

	// Check to see if the next level has already been created and create it if necessary
	if parent.Children[id[0]] == nil {
		parent.Children[id[0]] = NewNode()
	}

	// Recurse
	return db.setNodeRecurse(id[1:], parent.Children[id[0]], child)
}

func (db *Storage) removeNode(id NodeLocator) {
	// Find node
	node, path := db.findNode(id)
	b := id.GetBytes()

	if node != nil {
		parent := path[len(path)-1]
		// Node found, remove node and prune tree
		parent.Children[b[len(b)-1]] = nil
		db.prune(b[:len(b)-1], path[:len(path)-1])
	}
}

// Cleanup empty nodes in the given path
func (db *Storage) prune(id []byte, path []*Node) {
	if len(id) == 0 {
		// All done
		return
	}

	if len(id) > 1 && len(path) < 1 {
		// This shouldn't happen
		db.Logger.Printf("WARNING: prune() error - length of ID is %d but length of path is %d\n", len(id), len(path))
		return
	}

	parent := db.storage
	if len(path) > 0 {
		parent = path[len(path)-1]
	}

	nodeID := id[len(id)-1]
	node := parent.Children[nodeID]

	if node == nil || node.IsEmpty() {
		// Prune node
		//db.Logger.Printf("Prune: Removed node %X\n", id)
		parent.Children[nodeID] = nil
	}

	db.prune(id[:len(id)-1], path[:len(path)-1])

}

// Hash returns the 32bit hash for a given key
func Hash(key string) uint32 {
	h := fnv.New32()
	h.Write([]byte(key))
	return h.Sum32()
}

// Get returns the value of the given key
func (db *Storage) Get(id string) string {
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
func (db *Storage) Set(id string, value string) string {
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
func (db *Storage) Remove(id string) string {
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
func (db *Storage) Close() {
	db.Shutdown <- true
}
