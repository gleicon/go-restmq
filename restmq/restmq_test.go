// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package restmq

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var rmq *RestMQ

func init() {
	rmq = New("127.0.0.1:6379")
	rand.Seed(time.Now().UTC().UnixNano())
}

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

func TestAddAndGet(t *testing.T) {
	item1, err := rmq.Add("foobar-queue", "foobar-value")
	if err != nil {
		t.Error(err)
		return
	} else if item1["value"] != "foobar-value" {
		t.Error(fmt.Sprintf("Unexpected item's value: %#v", item1))
		return
	}
	item2, err := rmq.Get("foobar-queue", false)
	if err != nil {
		t.Error(err)
		return
	} else if item2["value"] != item1["value"] {
		t.Error(fmt.Sprintf("Unexpected item's value: %#v != %#v",
			item2["value"], item1["value"]))
		return
	}
}
