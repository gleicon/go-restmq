// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//
// RestMQ protocol in Go, based on
// https://github.com/gleicon/restmq/blob/master/src/restmq/core.py
//
// TODO: update the URL above when this branch turns master.
//
// RestMQ is controlled by a handler, obtained by calling the New function.
//
// The handler currently uses Redis as the backend, but there are plans to
// support leveldb and perhaps even other storage engines. (MariaDB?)
package restmq

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/fiorix/go-redis/redis"
)

// RestMQ is the instance's handler.
// It currently supports Redis but might become more modular, soon.
type RestMQ struct {
	Redis *redis.Client
}

// Item is the queue item.
type Item map[string]interface{}

// String returns the JSON-encoded representation of the Item.
func (it Item) String() string {
	if len(it) == 0 {
		return ""
	}
	b, err := json.Marshal(it)
	if err != nil {
		return ""
	}
	return string(b)
}

// New creates, initializes and returns a new instance of RestMQ.
// RestMQ instances are thread-safe.
func New(opts string) *RestMQ {
	return &RestMQ{redis.New(opts)}
}

// Add adds one item into the given queue, which is created on demand.
func (mq *RestMQ) Add(queue, value string) (Item, error) {
	item := make(Item)
	// TODO: Fix for cases when Redis disconnects in between the commands.
	uuid, err := mq.Redis.Incr(queue_uuid(queue))
	if err != nil {
		return item, err
	}
	lkey := queue + ":" + strconv.Itoa(uuid)
	if err := mq.Redis.Set(lkey, value); err != nil {
		return item, err
	}
	if _, err := mq.Redis.LPush(queue_name(queue), lkey); err != nil {
		return item, err
	}
	item["key"] = lkey
	item["value"] = value
	return item, nil
}

// Get returns one item from the given queue.
// Get can be "soft" or "hard" (hard is the default).
//
// A "soft" get is when an item is returned but remains in the queue, and
// its reference counter is incremented.
//
// The "hard" get returns an item from the queue without incrementing its
// reference counter. The item is permanently erased from the queue.
func (mq *RestMQ) Get(queue string, soft bool) (Item, error) {
	item := make(Item)
	var err error
	qn := queue_name(queue)
	var k string
	if soft {
		k, err = mq.Redis.LIndex(qn, -1)
	} else {
		k, err = mq.Redis.RPop(qn)
	}
	if err != nil {
		return item, err // Redis error
	} else if k == "" {
		return item, err // Empty queue
	}
	var v string
	v, err = mq.Redis.Get(k)
	if err != nil {
		return item, err // Redis error
	}
	item["key"] = k
	item["value"] = v
	return item, nil
}

// GetDel is the "hard" Get.
func (mq *RestMQ) GetDel(queue string) (string, error) {
	return "", nil
}

// Del does something.
// TODO: Gleicon ^^
func (mq *RestMQ) Del(queue string) error {
	return nil
}

// Len returns the length of the given queue.
func (mq *RestMQ) Len(queue string) (int, error) {
	return 0, nil
}

// All returns all from the given queue.
func (mq *RestMQ) All(queue string) ([]string, error) {
	return nil, nil
}

// Policy returns the current policy of the given queue.
func (mq *RestMQ) Policy(queue string) (string, error) {
	return "", nil
}

// PolicySet sets a new policy for the given queue.
func (mq *RestMQ) PolicySet(queue string) error {
	return nil
}

// Tail follows a queue.
// TODO: Gleicon update
// TODO: Channel?
func (mq *RestMQ) Tail(queue string) error {
	return nil
}

// CountElements does something different than Len.
// TODO: Gleicon
func (mq *RestMQ) CountElements(queue string) (int, error) {
	return 0, nil
}

// LastItems returns the last n items in the queue.
func (mq *RestMQ) LastItems(queue string, n int) ([]string, error) {
	return nil, nil
}

// TODO: add missing functions
