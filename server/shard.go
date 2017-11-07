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

package server

//noinspection GoRedundantImportAlias
import (
	"hash/fnv"

	"github.com/satori/go.uuid"
)

// Shard represents a single shard in the system that can contain some set of chunks
type Shard struct {
	ID      uuid.UUID
	Address string
}

// Chunk represents a single chunk in the system
type Chunk struct {
}

// ClusterSize is an enumeration for keeping track of how large a cluster is
type ClusterSize uint8

//noinspection GoUnusedGlobalVariable
var (
	// SmallCluster is for a cluster of up to 256 shards
	SmallCluster ClusterSize = 1
	// MediumCluster is for a cluster of up to 65,536 shards
	MediumCluster ClusterSize = 2
	// LargeCluster is for a cluster of up to 1,6777,216 shards
	LargeCluster ClusterSize = 3
	// HugeCluster is for a cluster of up to 4,294,967,296 shards
	HugeCluster ClusterSize = 4
)

// ClusterConfig holds the current cluster configuration
type ClusterConfig struct {
	Size ClusterSize
	// Chunks is the list of all chunks in the system
	Chunks []Chunk
	// Shards is the list of shards in the system
	Shards []Shard
}

// Chunk returns the 16bit chunk number for the given key.
func (config *ClusterConfig) Chunk(s string) uint32 {
	return uint32(Hash(s) >> (config.Size * 8))
}

// Hash returns the 32bit hash for a given key
func Hash(s string) uint32 {
	h := fnv.New32()
	h.Write([]byte(s))
	return h.Sum32()
}
