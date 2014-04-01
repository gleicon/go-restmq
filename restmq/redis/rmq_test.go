// Copyright 2010-2014 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rmq

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

var (
	testMQ   *Queue
	testOnce sync.Once
)

/*
func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
*/

// Tests

func startTestMQ() {
	testMQ = New("127.0.0.1:6379")
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestAdd(t *testing.T) {
	testOnce.Do(startTestMQ)

	item, err := testMQ.Add("myqueue", "foobar")
	if err != nil {
		t.Fatal(err)
	}

	if item.Value != "foobar" {
		t.Fatalf("Want 'foobar', have '%#v'", item.Value)
	}
}

func TestGet(t *testing.T) {
	item, err := testMQ.Get("myqueue", true) // pop
	if err != nil {
		t.Fatal(err)
	}

	if item.Value != "foobar" {
		t.Fatalf("Want 'foobar', have '%#v'", item.Value)
		return
	}
}

func BenchmarkAdd(b *testing.B) {
	var err error
	for n := 0; n < b.N; n++ {
		if _, err = testMQ.Add("myqueue", "foobar"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	testMQ.Add("myqueue", "foobar")
	defer testMQ.Get("myqueue", true)
	var err error
	for n := 0; n < b.N; n++ {
		if _, err = testMQ.Get("myqueue", false); err != nil {
			b.Fatal(err)
		}
	}
}
