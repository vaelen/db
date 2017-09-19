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
	"log"

	"encoding/binary"
	"hash/fnv"
)

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

// Hashtable implements a tree based hashtable
type Hashtable struct {
	root *Node
}

// NewHashtable creates a Hashtable instance
func NewHashtable() *Hashtable {
	return &Hashtable{
		root: NewNode(),
	}
}

// FindNode returns a node and the path to its parent node
func (db *Hashtable) FindNode(id NodeLocator) (*Node, []*Node) {
	path := make([]*Node, 0)
	return db.FindNodeRecurse(id.GetBytes(), db.root, path)
}

// FindNodeRecurse returns a node and the path to its parent node
func (db *Hashtable) FindNodeRecurse(id []byte, parent *Node, path []*Node) (*Node, []*Node) {
	if parent == nil {
		// We've hit a deadend
		return nil, path
	}

	if len(id) == 0 {
		// We found it
		return parent, path
	}

	// Not yet found, recurse
	return db.FindNodeRecurse(id[1:], parent.Children[id[0]], append(path, parent))
}

// SetNodeValue sets a key/value pair on a given node
func (db *Hashtable) SetNodeValue(id NodeLocator, key string, value string) *Node {
	node, _ := db.FindNode(id)
	if node != nil {
		node.SetValue(key, value)
	} else {
		node = NewNode()
		node.SetValue(key, value)
		db.SetNode(id, node)
	}
	return node
}

// SetNode sets a given node in the tree
func (db *Hashtable) SetNode(id NodeLocator, value *Node) *Node {
	return db.setNodeRecurse(id.GetBytes(), db.root, value)
}

func (db *Hashtable) setNodeRecurse(id []byte, parent *Node, child *Node) *Node {
	if parent == nil || len(id) == 0 {
		// This should never happen, but just in case...
		log.Printf("WARNING: setNodeRecurse() called with either no parent or no id.\n")
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

// RemoveNode removes a given node from the tree
func (db *Hashtable) RemoveNode(id NodeLocator) {
	// Find node
	node, path := db.FindNode(id)
	b := id.GetBytes()

	if node != nil {
		parent := path[len(path)-1]
		// Node found, remove node and prune tree
		parent.Children[b[len(b)-1]] = nil
		db.prune(b[:len(b)-1], path[:len(path)-1])
	}
}

// Cleanup empty nodes in the given path
func (db *Hashtable) prune(id []byte, path []*Node) {
	if len(id) == 0 {
		// All done
		return
	}

	if len(id) > 1 && len(path) < 1 {
		// This shouldn't happen
		log.Printf("WARNING: prune() error - length of ID is %d but length of path is %d\n", len(id), len(path))
		return
	}

	parent := db.root
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

// Get returns the value for a given key
func (db *Hashtable) Get(key string) string {
	id := GetNodeLocator(key)
	node, _ := db.FindNode(id)
	if node != nil {
		return node.GetValue(key)
	}
	return ""
}

// Set sets the value for a given key
func (db *Hashtable) Set(key string, value string) {
	id := GetNodeLocator(key)
	node := db.SetNodeValue(id, key, value)
	if node != nil {
		node.SetValue(key, value)
	}
}

// Remove removes a given key
func (db *Hashtable) Remove(key string) {
	id := GetNodeLocator(key)
	node, path := db.FindNode(id)
	if node != nil {
		node.RemoveValue(key)
		db.prune(id.GetBytes(), path)
	}
}
