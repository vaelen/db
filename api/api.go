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

//go:generate protoc vdb.proto --go_out=plugins=grpc:.

package api

import (
	"io"
	"bytes"
	"encoding/binary"
)

// Multiplexer multiplexes byte arrays to and from a ReaderWriter
type Multiplexer struct {
	Stream io.ReadWriter
}

// Send writes a message size and a message to the writer
func (m *Multiplexer) Send(message []byte) error {
	// Write message size
	buf := bytes.Buffer{}

	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes[0:], uint64(len(message)))
	buf.Write(sizeBytes)

	// Write message
	buf.Write(message)

	_, err := m.Stream.Write(buf.Bytes())
	return err
}

// Receive reads a message size and a message from the reader
func (m *Multiplexer) Receive() ([]byte, error) {
	buf := bytes.Buffer{}

	// Read message size
	n, err := buf.ReadFrom(io.LimitReader(m.Stream, 8))
	if err != nil {
		return nil, err
	}
	if n < 8 {
		return nil, io.EOF
	}

	size := int64(binary.BigEndian.Uint64(buf.Bytes()))

	buf.Reset()

	// Read message
	n, err = buf.ReadFrom(io.LimitReader(m.Stream, size))
	if err != nil {
		return buf.Bytes(), err
	}
	if n < size {
		return buf.Bytes(), io.EOF
	}

	return buf.Bytes(), nil
}