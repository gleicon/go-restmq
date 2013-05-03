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
	"github.com/fiorix/go-redis/redis"
)

// RestMQ is the instance's handler.
// It currently supports Redis but might become more modular, soon.
type RestMQ struct {
	Redis *redis.Client
}

var QUEUESET = "QUEUESET"
var UUID_SUFFIX = ":UUID"
var QUEUE_SUFFIX = ":queue"

// New creates, initializes and returns a new instance of RestMQ.
// RestMQ instances are thread-safe.
func New(opts string) *RestMQ {
	return &RestMQ{redis.New(opts)}
}

// Add adds one item into the given queue, which is created on demand.
func (mq *RestMQ) Add(queue, items string) error {
	return nil
}

// Get returns one item from the given queue.
// Get can be "soft" or "hard" (hard is the default).
//
// A "soft" get is when an item is returned but remains in the queue, and
// its reference counter is incremented.
//
// The "hard" get returns an item from the queue without incrementing its
// reference counter. The item is permanently erased from the queue.
func (mq *RestMQ) Get(queue string, softget bool) (string, error) {
	return "", nil
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
