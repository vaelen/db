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
	"math/rand"
	"os"
	"testing"
	"time"
)

func randomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// TestStorage tests a Storage instance
func TestStorage(t *testing.T) {
	t.Logf("Testing Storage\n")
	
	// Generate some random data
	rand.Seed(time.Now().UnixNano())

	t.Logf("Generating random strings\n")

	m := make(map[string]string)
	for i := 0; i < 100000; i++ {
		m[randomString(1000)] = randomString(1000)
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Create a storage instance
	s := New(os.Stderr, "")
	defer s.Close()

	t.Logf("Adding random strings\n")
	for k, v := range m {
		s.Set(k, v)
	}

	t.Logf("Modifying values\n")
	for i := 0; i < 1000; i++ {
		key := keys[i]
		newValue := randomString(1000)
		m[key] = newValue
		s.Set(key, newValue)
	}

	t.Logf("Getting random strings\n")
	for k, v := range m {
		x := s.Get(k)
		if x != v {
			t.Errorf("Error Getting Value: Key: %s\n", k)
		}
	}

	t.Logf("Removing random strings\n")
	for k := range m {
		_ = s.Remove(k)
	}

}


// TestHashtable tests a Hashtable instance
func TestHashtable(t *testing.T) {
	t.Logf("Testing Hashtable\n")
	
	// Generate some random data
	rand.Seed(time.Now().UnixNano())

	t.Logf("Generating random strings\n")

	m := make(map[string]string)
	for i := 0; i < 100000; i++ {
		m[randomString(1000)] = randomString(1000)
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Create a Hashtable instance
	h := NewHashtable()

	t.Logf("Adding random strings\n")
	for k, v := range m {
		h.Set(k, v)
	}

	t.Logf("Modifying values\n")
	for i := 0; i < 1000; i++ {
		key := keys[i]
		newValue := randomString(1000)
		m[key] = newValue
		h.Set(key, newValue)
	}

	t.Logf("Getting random strings\n")
	for k, v := range m {
		x := h.Get(k)
		if x != v {
			t.Errorf("Error Getting Value: Key: %s\n", k)
		}
	}

	t.Logf("Removing random strings\n")
	for k := range m {
		h.Remove(k)
	}

}
